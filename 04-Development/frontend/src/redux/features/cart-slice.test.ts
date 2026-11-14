import { describe, expect, it } from "vitest";
import { configureStore } from "@reduxjs/toolkit";
import reducer, { clearCart, selectCart, selectCartCount, selectCartItems, setCart } from "./cart-slice";
import type { ApiCart, ApiCartItem } from "@/types/api";

const item = (id: string, quantity: number) => ({ id, quantity }) as unknown as ApiCartItem;
const makeStore = () => configureStore({ reducer: { cartReducer: reducer } });

describe("cart slice", () => {
  it("state awal kosong dan belum dimuat", () => {
    expect(reducer(undefined, { type: "@@init" })).toEqual({ items: [], subtotal: "0", loaded: false });
  });

  it("setCart menyalin items dan subtotal, loaded = true", () => {
    const cart = { items: [item("a", 1)], subtotal: "150000" } as unknown as ApiCart;
    const s = reducer(undefined, setCart(cart));
    expect(s).toEqual({ items: cart.items, subtotal: "150000", loaded: true });
  });

  it("clearCart mengosongkan tetapi tetap loaded", () => {
    const filled = reducer(undefined, setCart({ items: [item("a", 1)], subtotal: "1" } as unknown as ApiCart));
    expect(reducer(filled, clearCart())).toEqual({ items: [], subtotal: "0", loaded: true });
  });

  it("selector: items, cart, dan jumlah kuantitas", () => {
    const store = makeStore();
    expect(selectCartCount(store.getState())).toBe(0);
    store.dispatch(setCart({ items: [item("a", 2), item("b", 3)], subtotal: "5" } as unknown as ApiCart));
    const st = store.getState();
    expect(selectCartItems(st)).toHaveLength(2);
    expect(selectCart(st).subtotal).toBe("5");
    expect(selectCartCount(st)).toBe(5);
  });
});
