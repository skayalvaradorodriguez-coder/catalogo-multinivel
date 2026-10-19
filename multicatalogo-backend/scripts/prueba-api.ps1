<#
  prueba-api.ps1 — Prueba de humo del back-office de MultiCatálogo contra la API real.

  Hace por comandos lo mismo que se revisa en el navegador:
    1. Cliente: Mi Red (link de invitación) y Mis Comisiones
    2. Cliente: compra un producto y consulta Mis Pedidos + detalle
    3. Admin:   Dashboard (reportes), productos, pedidos, usuarios y liquidaciones pendientes

  Uso (con el backend corriendo):
    cd multicatalogo-backend
    powershell -ExecutionPolicy Bypass -File scripts\prueba-api.ps1 -Email tu@correo.com

  La contraseña se pide por teclado si no se pasa con -Password.
#>
param(
    [string]$Api = "http://127.0.0.1:3000",
    [string]$Email = "",
    [string]$Password = "",
    [string]$AdminEmail = "admin@upse.edu.ec",
    [string]$AdminPassword = "123456",
    [int]$ProductoId = 2
)

$ErrorActionPreference = "Stop"
$script:fallas = 0

function Titulo($texto) { Write-Host ""; Write-Host "=== $texto ===" -ForegroundColor Cyan }
function Ok($texto) { Write-Host "  [OK] $texto" -ForegroundColor Green }
function Falla($texto) { Write-Host "  [FALLA] $texto" -ForegroundColor Red; $script:fallas++ }
function Dinero($valor) { return ('$' + ([double]$valor).ToString("0.00")) }

# Llama a la API. Devuelve el JSON ya convertido en objeto; si la API responde con error,
# muestra el mensaje del APIError y devuelve $null.
function Api($metodo, $ruta, $token = $null, $cuerpo = $null) {
    $params = @{ Method = $metodo; Uri = "$Api$ruta"; Headers = @{} }
    if ($token) { $params.Headers["Authorization"] = "Bearer $token" }
    if ($null -ne $cuerpo) {
        $params.ContentType = "application/json; charset=utf-8"
        # Bytes UTF-8 para que las tildes lleguen bien también en Windows PowerShell 5.1.
        $params.Body = [System.Text.Encoding]::UTF8.GetBytes(($cuerpo | ConvertTo-Json -Depth 6 -Compress))
    }
    try {
        return Invoke-RestMethod @params
    } catch {
        $detalle = $_.ErrorDetails.Message
        if (-not $detalle) { $detalle = $_.Exception.Message }
        Falla "$metodo $ruta -> $detalle"
        return $null
    }
}

function IniciarSesion($correo, $clave) {
    $r = Api "POST" "/api/login" $null @{ email = $correo; password = $clave }
    if ($r) { Ok "Sesión iniciada como $($r.email) (rol $($r.rol))"; return $r.token }
    return $null
}

# ---------------------------------------------------------------------------------------
Titulo "Backend"
$salud = Api "GET" "/api/health"
if (-not $salud) { Write-Host "El backend no responde en $Api. Arráncalo con: go run ." -ForegroundColor Yellow; exit 1 }
Ok "health: $($salud.status) (postgres $($salud.postgres), redis $($salud.redis))"

if (-not $Email) { $Email = Read-Host "Email de tu cuenta de cliente" }
if (-not $Password) {
    $seguro = Read-Host "Contraseña de $Email" -AsSecureString
    $Password = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($seguro))
}

# ---------------------------------------------------------------------------------------
Titulo "1. Cliente: Mi Red y Mis Comisiones"
$token = IniciarSesion $Email $Password
if (-not $token) { Write-Host "No se pudo iniciar sesión con $Email; revisa el email y la contraseña." -ForegroundColor Yellow; exit 1 }

$red = Api "GET" "/api/red/mia" $token
if ($red) {
    Ok "Nodo en la red: $($red.nombre) · código $($red.codigo)"
    Ok "Link de invitación: http://localhost:5173/registro?ref=$($red.codigo)"
    $directos = @($red.hijos).Where({ $_ }).Count
    Ok "Referidos directos: $directos"
}
$cuenta = Api "GET" "/api/comisiones/mias" $token
if ($cuenta) {
    Ok ("Mis comisiones -> por cobrar {0} · cobrado {1} · revertido {2} · movimientos {3}" -f `
        (Dinero $cuenta.totales.pendiente), (Dinero $cuenta.totales.pagada), (Dinero $cuenta.totales.revertida), @($cuenta.comisiones).Count)
}

# ---------------------------------------------------------------------------------------
Titulo "2. Cliente: compra y Mis Pedidos"
$producto = Api "GET" "/api/productos/$ProductoId"
$pedido = $null
if ($producto) {
    Ok "Producto a comprar: $($producto.nombre) · $(Dinero $producto.precio) · stock $($producto.stock)"
    $pedido = Api "POST" "/api/checkout" $token @{
        cliente    = @{ nombre = "Prueba por comandos"; email = $Email; direccion = "Av. Principal 123"; ciudad = "La Libertad"; telefono = "0999999999" }
        metodoPago = "Efectivo contra entrega"
        items      = @(@{ id = $ProductoId; cantidad = 1 })
    }
}
if ($pedido) {
    Ok "Pedido creado: $($pedido.numero) · total $(Dinero $pedido.total) · estado $($pedido.estado)"
    Start-Sleep -Milliseconds 500  # la goroutine de comisiones corre en segundo plano

    $mios = Api "GET" "/api/pedidos/mios?page=1&limit=5" $token
    if ($mios) {
        $esta = @($mios.data | Where-Object { $_.id -eq $pedido.id }).Count -eq 1
        if ($esta) { Ok "Mis Pedidos: $($mios.total) pedido(s); el nuevo aparece en la lista" } else { Falla "El pedido nuevo no aparece en Mis Pedidos" }
        foreach ($p in $mios.data) { Write-Host ("       {0}  {1,-10} {2,9}  {3} u." -f $p.numero, $p.estado, (Dinero $p.total), $p.unidades) }
    }
    $detalle = Api "GET" "/api/pedidos/$($pedido.id)" $token
    if ($detalle) {
        Ok "Detalle: envío a $($detalle.cliente.direccion), $($detalle.cliente.ciudad) · pago $($detalle.metodoPago)"
        foreach ($i in $detalle.items) { Ok "  $($i.nombre) x $($i.cantidad) = $(Dinero ($i.precio * $i.cantidad))" }
    }
}

# ---------------------------------------------------------------------------------------
Titulo "3. Admin: Dashboard, gestión y liquidaciones"
$admin = IniciarSesion $AdminEmail $AdminPassword
if ($admin) {
    $resumen = Api "GET" "/api/reportes/resumen" $admin
    if ($resumen) {
        Ok ("Dashboard -> vendido {0} · pedidos {1} · ticket {2} · comisiones devengadas {3}" -f `
            (Dinero $resumen.totalVendido), $resumen.numPedidos, (Dinero $resumen.ticketPromedio), (Dinero $resumen.comisionesDevengadas))
    }
    $top = Api "GET" "/api/reportes/top-productos?limit=3" $admin
    if ($top) { Ok ("Top productos: " + (($top | ForEach-Object { "$($_.nombre) ($($_.unidades) u.)" }) -join ", ")) }

    $prods = Api "GET" "/api/admin/productos?page=1&limit=1" $admin
    if ($prods) { Ok "Productos (admin): $($prods.total) en total, incluidos los desactivados" }
    $peds = Api "GET" "/api/pedidos?page=1&limit=1" $admin
    if ($peds) { Ok "Pedidos (admin): $($peds.total) en total" }
    $usrs = Api "GET" "/api/usuarios?page=1&limit=1" $admin
    if ($usrs) { Ok "Usuarios (admin): $($usrs.total) cuentas" }

    $deudas = Api "GET" "/api/reportes/comisiones-pendientes" $admin
    if ($null -ne $deudas) {
        if (@($deudas).Count -eq 0) { Ok "Liquidaciones: no hay comisiones pendientes" }
        foreach ($d in $deudas) { Ok "Liquidaciones -> $($d.nombre) ($($d.codigo)) tiene $(Dinero $d.pendiente) por cobrar en $($d.comisiones) comisión(es)" }
        if ($red -and $red.codigo) {
            # Quién debe cobrar por la compra de este usuario: su patrocinador directo.
            Write-Host "  (La compra de $Email paga 10% a su patrocinador directo, 5% al siguiente y 2% al tercero)" -ForegroundColor DarkGray
        }
    }

    # Seguridad: un cliente no puede entrar a rutas de administrador.
    $params = @{ Method = "GET"; Uri = "$Api/api/reportes/resumen"; Headers = @{ Authorization = "Bearer $token" } }
    try { Invoke-RestMethod @params | Out-Null; Falla "Un cliente pudo ver los reportes del admin" }
    catch { if ($_.Exception.Response.StatusCode.value__ -eq 403) { Ok "Seguridad: el cliente recibe 403 en rutas de admin" } else { Falla "Se esperaba 403: $($_.Exception.Message)" } }
}

# ---------------------------------------------------------------------------------------
Titulo "Resultado"
if ($script:fallas -eq 0) { Write-Host "  Todas las pruebas pasaron." -ForegroundColor Green }
else { Write-Host "  $($script:fallas) prueba(s) fallaron; revisa los mensajes en rojo." -ForegroundColor Red; exit 1 }
