"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Search } from "lucide-react";
import { PageHead, btnLine, inputCls, stackTable, theadCls } from "@/components/admin/ui";
import { CUST_STATUS, TABS } from "@/components/order/status";
import { Badge, EmptyState, Pagination } from "@/components/ui/misc";
import { errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { formatDate, formatRupiah, paymentLabel } from "@/lib/format";
import { adminService } from "@/services";
import type { ApiOrderListItem } from "@/types/api";

const LIMIT = 20;

export default function AdminPesananPage() {
  const [tab, setTab] = useState(0);
  const [q, setQ] = useState("");
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<ApiOrderListItem[] | null>(null);
  const [total, setTotal] = useState(0);
  const match = TABS[tab].match;
  const multi = !!match && match.length > 1;

  useEffect(() => {
    const t = setTimeout(() => {
      const query = { q: q.trim() || undefined };
      // API hanya menerima satu order_status; tab gabungan (paid + processing) diambil terpisah lalu digabung.
      // shortcut: tab gabungan tanpa paginasi (maks 100 per status), tambah filter multi-status di API bila perlu
      const load = multi
        ? Promise.all(match!.map((s) => adminService.orders({ ...query, order_status: s, limit: 100 }))).then((rs) => {
            const data = rs.flatMap((r) => r.data).sort((a, b) => b.created_at.localeCompare(a.created_at));
            return { data, total: data.length };
          })
        : adminService.orders({ ...query, order_status: match?.[0], page, limit: LIMIT }).then((r) => ({ data: r.data, total: r.meta.total }));
      load
        .then((r) => {
          setRows(r.data);
          setTotal(r.total);
        })
        .catch((e) => {
          setRows([]);
          toast.error(errorMessage(e));
        });
    }, 300);
    return () => clearTimeout(t);
  }, [tab, q, page, multi, match]);

  return (
    <div>
      <PageHead title="Pesanan Masuk" sub="Pesanan dari website Hii Stranger." />
      <div className="mb-4 flex gap-3 overflow-x-auto pb-1" role="tablist" aria-label="Filter status pesanan">
        {TABS.map((t, i) => (
          <button
            key={t.label}
            role="tab"
            aria-selected={i === tab}
            onClick={() => {
              setTab(i);
              setPage(1);
            }}
            className={cn(
              "shrink-0 rounded-full border px-6 py-3 text-lg font-bold",
              i === tab ? "border-ink bg-ink text-white" : "border-line-strong bg-white",
            )}
          >
            {t.label}
          </button>
        ))}
      </div>
      <label className="relative mb-6 block">
        <span className="sr-only">Cari nomor pesanan</span>
        <Search className="absolute left-4 top-1/2 size-4 -translate-y-1/2 text-faint" aria-hidden />
        <input
          value={q}
          onChange={(e) => {
            setQ(e.target.value);
            setPage(1);
          }}
          placeholder="Cari nomor pesanan"
          className={inputCls + " h-12 pl-11 text-base"}
        />
      </label>
      <div className="overflow-x-auto rounded-xl border border-line-strong bg-white">
        <table className={"w-full text-xs sm:min-w-[900px] " + stackTable}>
          <thead className={theadCls}>
            <tr>
              {["No.", "Tanggal", "Pelanggan", "Total", "Pembayaran", "Status", "Tujuan", ""].map((h, i) => (
                <th key={i} className="px-5 py-4 font-semibold">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows?.map((o) => {
              const pay = paymentLabel(o);
              const st = CUST_STATUS[o.order_status];
              return (
                <tr key={o.id} className="border-t border-line-strong">
                  <td className="px-5 py-3">#{o.order_number}</td>
                  <td data-l="Tanggal" className="px-5 py-3">{formatDate(o.created_at)}</td>
                  <td data-l="Pelanggan" className="px-5 py-3 font-semibold">{o.recipient_name}</td>
                  <td data-l="Total" className="px-5 py-3">{formatRupiah(o.total_amount)}</td>
                  <td data-l="Pembayaran" className="px-5 py-3">
                    <Badge tone={pay.tone} className="px-2 py-0.5 text-[10px]">
                      {pay.label}
                    </Badge>
                  </td>
                  <td data-l="Status" className="px-5 py-3">
                    <span className={cn("rounded-full px-2 py-0.5 text-[10px] font-bold", st?.cls)}>{st?.label ?? o.order_status}</span>
                  </td>
                  <td data-l="Tujuan" className="px-5 py-3">{o.shipping_city ?? "-"}</td>
                  <td className="px-5 py-3 text-right">
                    <Link href={"/admin/pesanan/" + o.id} className={btnLine}>
                      Detail
                    </Link>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {rows?.length === 0 && <EmptyState title="Tidak ada pesanan" className="py-10" />}
      </div>
      {!multi && (
        <div className="mt-4">
          <Pagination page={page} limit={LIMIT} total={total} onPage={setPage} />
        </div>
      )}
    </div>
  );
}
