import { useState } from "react";

// Link de invitación del usuario con botón para copiarlo.
const LinkReferido = ({ codigo }) => {
  const [copiado, setCopiado] = useState(false);
  const link = `${window.location.origin}/registro?ref=${codigo}`;

  const copiar = async () => {
    try {
      await navigator.clipboard.writeText(link);
      setCopiado(true);
      setTimeout(() => setCopiado(false), 2000);
    } catch {
      window.prompt("Copia tu link de invitación:", link);
    }
  };

  return (
    <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-4 mb-6">
      <p className="text-sm font-semibold text-indigo-900 mb-2">
        Tu link de invitación · código <span className="font-mono">{codigo}</span>
      </p>
      <div className="flex flex-col sm:flex-row gap-2">
        <input
          readOnly
          value={link}
          onFocus={(e) => e.target.select()}
          className="flex-1 min-w-0 px-3 py-2 rounded-lg border border-indigo-200 bg-white text-sm font-mono text-slate-700"
        />
        <button
          onClick={copiar}
          className="min-h-11 px-5 rounded-lg bg-indigo-600 text-white font-semibold hover:bg-indigo-700 transition"
        >
          {copiado ? "¡Copiado!" : "Copiar link"}
        </button>
      </div>
    </div>
  );
};

export default LinkReferido;
