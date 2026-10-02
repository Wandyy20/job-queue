import { useState, useCallback } from "react";
import ToastContext from "../context/ToastContext";

export default function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const showToast = useCallback((message, type = "success") => {
    const id = Date.now();
    setToasts((t) => [...t, { id, message, type }]);
    setTimeout(() => {
      setToasts((t) => t.filter((toast) => toast.id !== id));
    }, 3000);
  }, []);

  return (
    <ToastContext.Provider value={showToast}>
      {children}
      <div className="fixed bottom-4 right-4 space-y-2 z-50">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`px-4 py-2.5 rounded-md text-sm shadow-lg border ${
              t.type === "error"
                ? "bg-[#1A1114] border-[#F87171]/30 text-[#F87171]"
                : "bg-[#0F1A16] border-[#34D399]/30 text-[#34D399]"
            }`}
          >
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}