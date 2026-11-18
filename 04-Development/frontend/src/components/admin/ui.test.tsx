import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AvailBadge } from "./ui";

describe("AvailBadge", () => {
  it("ready stock tidak menampilkan estimasi walau data lama masih punya hari", () => {
    const { container } = render(<AvailBadge type="ready" days={7} />);
    expect(container.textContent).toBe("Ready Stock");
  });

  it("pre-order menampilkan estimasi hari", () => {
    render(<AvailBadge type="preorder" days={7} />);
    expect(screen.getByText("Pre-Order · 7 hari")).toBeTruthy();
  });
});
