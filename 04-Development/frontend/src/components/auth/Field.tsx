"use client";

import { Eye, EyeOff, Lock, TriangleAlert, User } from "lucide-react";
import { useId, useState, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

type Props = InputHTMLAttributes<HTMLInputElement> & { label: string; error?: string; icon?: "user" | "lock" };

export default function Field({ label, error, icon, type, className, ...rest }: Props) {
  const id = useId();
  const [show, setShow] = useState(false);
  const Icon = icon === "lock" ? Lock : icon === "user" ? User : null;
  const pw = type === "password";
  return (
    <div>
      <label htmlFor={id} className="mb-2 block font-semibold">
        {label} {rest.required && <span className="text-danger">*</span>}
      </label>
      <div className="relative">
        {Icon && <Icon className="absolute left-4 top-1/2 size-5 -translate-y-1/2 text-faint" aria-hidden />}
        <input
          id={id}
          type={pw && show ? "text" : type}
          aria-invalid={!!error}
          className={cn(
            "h-15 w-full rounded-xl border bg-white px-6 focus:outline-2 focus:outline-ink",
            Icon && "pl-12",
            pw && "pr-12",
            error ? "border-danger bg-[#fff8f7]" : "border-faint",
            className,
          )}
          {...rest}
        />
        {pw && (
          <button
            type="button"
            onClick={() => setShow((s) => !s)}
            aria-label={show ? "Sembunyikan kata sandi" : "Tampilkan kata sandi"}
            className="absolute right-2 top-1/2 -translate-y-1/2 p-2 text-muted"
          >
            {show ? <EyeOff className="size-5" /> : <Eye className="size-5" />}
          </button>
        )}
      </div>
      {error && (
        <p role="alert" className="mt-1 flex items-center gap-1 text-xs text-danger">
          <TriangleAlert className="size-3" /> {error}
        </p>
      )}
    </div>
  );
}
