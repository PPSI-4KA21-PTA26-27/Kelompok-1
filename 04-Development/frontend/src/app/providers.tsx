"use client";
import React from "react";
import { Toaster } from "react-hot-toast";
import { ReduxProvider } from "@/redux/provider";
import { AuthProvider } from "@/app/context/AuthContext";
import CartSync from "@/app/CartSync";
import ServerGate from "@/components/layout/ServerGate";

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ReduxProvider>
      <ServerGate>
      <AuthProvider>
          <CartSync />
          {children}
          <Toaster
            position="top-center"
            toastOptions={{
              duration: 3200,
              style: {
                background: "#141413",
                color: "#fbf9f5",
                borderRadius: "12px",
                fontSize: "14px",
                padding: "10px 14px",
              },
              success: { iconTheme: { primary: "#d4ff14", secondary: "#141413" } },
              error: { iconTheme: { primary: "#ff6b5e", secondary: "#141413" } },
            }}
          />
      </AuthProvider>
    </ServerGate>
    </ReduxProvider>
  );
}
