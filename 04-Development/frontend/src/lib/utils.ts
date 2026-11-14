import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Utilitas className untuk komponen shadcn/ui (components.json -> aliases.utils).
 * Komponen lama tetap memakai `@/lib/cn`.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
