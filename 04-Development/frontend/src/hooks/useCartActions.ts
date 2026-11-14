"use client";
import { useCallback } from "react";
import { useRouter } from "next/navigation";
import toast from "react-hot-toast";
import { useAuth } from "@/app/context/AuthContext";
import { useAppDispatch } from "@/redux/store";
import { clearCart, setCart } from "@/redux/features/cart-slice";
import { cartService } from "@/services";
import { errorMessage } from "@/lib/api";

/** Operasi keranjang lewat backend (/cart). Hanya untuk pelanggan yang login. */
export function useCartActions() {
  const { user, isStaff } = useAuth();
  const dispatch = useAppDispatch();
  const router = useRouter();

  const refresh = useCallback(async () => {
    if (!user || isStaff) {
      dispatch(clearCart());
      return;
    }
    dispatch(setCart(await cartService.get()));
  }, [user, isStaff, dispatch]);

  /** Pastikan pemanggil boleh berbelanja; arahkan ke login bila belum masuk. */
  const guard = useCallback(
    (next?: string) => {
      if (!user) {
        toast("Masuk dulu untuk menyimpan ke tas belanja.");
        router.push("/signin" + (next ? "?next=" + encodeURIComponent(next) : ""));
        return false;
      }
      if (isStaff) {
        toast.error("Akun staf tidak dapat berbelanja. Gunakan akun pelanggan.");
        return false;
      }
      return true;
    },
    [user, isStaff, router]
  );

  const add = useCallback(
    async (variantId: string, quantity = 1, next?: string) => {
      if (!guard(next)) return false;
      try {
        await cartService.add(variantId, quantity);
        await refresh();
        return true;
      } catch (e) {
        toast.error(errorMessage(e));
        return false;
      }
    },
    [guard, refresh]
  );

  const updateQuantity = useCallback(
    async (cartItemId: string, quantity: number) => {
      try {
        await cartService.update(cartItemId, quantity);
      } catch (e) {
        toast.error(errorMessage(e));
      }
      await refresh().catch(() => {});
    },
    [refresh]
  );

  const removeItem = useCallback(
    async (cartItemId: string) => {
      try {
        await cartService.remove(cartItemId);
        await refresh();
      } catch (e) {
        toast.error(errorMessage(e));
      }
    },
    [refresh]
  );

  return { add, updateQuantity, removeItem, refresh, guard };
}
