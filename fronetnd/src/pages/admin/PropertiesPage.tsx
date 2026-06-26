import { useEffect, useState } from "react";
import { getProperties, updateProperty, type Property } from "@/lib/api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/components/ui/toaster";
import { Loader2, Pencil } from "lucide-react";

const COLOR_DOT: Record<string, string> = {
  brown:      "bg-amber-900",
  light_blue: "bg-sky-400",
  pink:       "bg-pink-400",
  orange:     "bg-orange-400",
  red:        "bg-red-500",
  yellow:     "bg-yellow-400",
  green:      "bg-green-500",
  dark_blue:  "bg-blue-700",
  railroad:   "bg-zinc-600",
  utility:    "bg-teal-500",
};

function rentLabel(p: Property) {
  if (p.color_group === "railroad") return ["1rr","2rr","3rr","4rr"];
  if (p.color_group === "utility")  return ["×1 owned","×2 owned"];
  return ["Base","1H","2H","3H","4H","Hotel"];
}

export default function PropertiesPage() {
  const [props, setProps] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<Property | null>(null);
  const [saving, setSaving] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    getProperties()
      .then(setProps)
      .catch(() => toast({ title: "Failed to load properties", variant: "destructive" }))
      .finally(() => setLoading(false));
  }, []);

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    try {
      await updateProperty(editing.id, {
        price:         editing.price,
        mortgage_value: editing.mortgage_value,
        house_cost:    editing.house_cost,
        rent:          editing.rent,
      });
      setProps((ps) => ps.map((p) => (p.id === editing.id ? editing : p)));
      toast({ title: "Property saved" });
      setEditing(null);
    } catch {
      toast({ title: "Failed to save", variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  const setRent = (i: number, val: number) =>
    setEditing((e) => {
      if (!e) return e;
      const rent = [...e.rent];
      rent[i] = val;
      return { ...e, rent };
    });

  return (
    <div className="p-6 space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Properties</h1>
        <p className="text-muted-foreground text-sm mt-1">
          All 28 ownable tiles — click to edit prices and rent table
        </p>
      </div>

      {loading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="rounded-xl border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8">Color</TableHead>
                <TableHead>Name</TableHead>
                <TableHead className="text-right">Price</TableHead>
                <TableHead className="text-right">Mortgage</TableHead>
                <TableHead className="text-right">House cost</TableHead>
                <TableHead className="text-right">Base rent</TableHead>
                <TableHead className="w-10"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {props.map((p) => (
                <TableRow key={p.id} className="cursor-pointer" onClick={() => setEditing({ ...p, rent: [...p.rent] })}>
                  <TableCell>
                    <span className={`inline-block h-4 w-4 rounded-sm ${COLOR_DOT[p.color_group] ?? "bg-zinc-400"}`} />
                  </TableCell>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell className="text-right font-mono">${p.price}</TableCell>
                  <TableCell className="text-right font-mono">${p.mortgage_value}</TableCell>
                  <TableCell className="text-right font-mono">
                    {p.house_cost ? `$${p.house_cost}` : <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell className="text-right font-mono">${p.rent?.[0] ?? 0}</TableCell>
                  <TableCell>
                    <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <Dialog open={!!editing} onOpenChange={(o) => !o && setEditing(null)}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>
              Edit — {editing?.name}
              {editing?.color_group && (
                <span className={`ml-2 inline-block h-3 w-3 rounded-sm align-middle ${COLOR_DOT[editing.color_group] ?? "bg-zinc-400"}`} />
              )}
            </DialogTitle>
          </DialogHeader>

          {editing && (
            <div className="space-y-4 py-2">
              <div className="grid grid-cols-3 gap-3">
                <div className="space-y-1.5">
                  <Label>Price</Label>
                  <Input type="number" value={editing.price}
                    onChange={(e) => setEditing({ ...editing, price: +e.target.value })} />
                </div>
                <div className="space-y-1.5">
                  <Label>Mortgage value</Label>
                  <Input type="number" value={editing.mortgage_value}
                    onChange={(e) => setEditing({ ...editing, mortgage_value: +e.target.value })} />
                </div>
                <div className="space-y-1.5">
                  <Label>House / upgrade cost</Label>
                  <Input type="number" value={editing.house_cost}
                    onChange={(e) => setEditing({ ...editing, house_cost: +e.target.value })} />
                </div>
              </div>

              <div className="space-y-2">
                <Label>Rent table</Label>
                <div className="grid gap-2">
                  {rentLabel(editing).map((lbl, i) => (
                    <div key={i} className="flex items-center gap-3">
                      <span className="w-24 text-xs text-muted-foreground shrink-0">{lbl}</span>
                      <Input
                        type="number"
                        value={editing.rent?.[i] ?? 0}
                        onChange={(e) => setRent(i, +e.target.value)}
                      />
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditing(null)}>Cancel</Button>
            <Button onClick={save} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 animate-spin" />}
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
