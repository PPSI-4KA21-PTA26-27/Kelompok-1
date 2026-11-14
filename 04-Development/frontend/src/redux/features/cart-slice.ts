import { createSelector, createSlice, PayloadAction } from "@reduxjs/toolkit";
import type { RootState } from "../store";
import type { ApiCart, ApiCartItem } from "@/types/api";

/**
 * Cermin keranjang dari backend (GET /cart). Perubahan dilakukan lewat
 * useCartActions yang memanggil API lalu menyimpan hasilnya di sini.
 */
type CartState = { items: ApiCartItem[]; subtotal: string; loaded: boolean };

const initialState: CartState = { items: [], subtotal: "0", loaded: false };

export const cart = createSlice({
  name: "cart",
  initialState,
  reducers: {
    setCart: (state, action: PayloadAction<ApiCart>) => {
      state.items = action.payload.items;
      state.subtotal = action.payload.subtotal;
      state.loaded = true;
    },
    clearCart: (state) => {
      state.items = [];
      state.subtotal = "0";
      state.loaded = true;
    },
  },
});

export const selectCart = (state: RootState) => state.cartReducer;
export const selectCartItems = (state: RootState) => state.cartReducer.items;
export const selectCartCount = createSelector([selectCartItems], (items) =>
  items.reduce((n, it) => n + it.quantity, 0)
);

export const { setCart, clearCart } = cart.actions;
export default cart.reducer;
