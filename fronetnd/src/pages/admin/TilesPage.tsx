import { useEffect, useState } from "react";
import { getTiles, updateTile, type Tile } from "@/lib/api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/components/ui/toaster";
import { Loader2, Pencil } from "lucide-react";

const COLOR_MAP: Record<string, string> = {
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

const TYPE_VARIANT: Record<string, "default" | "secondary" | "outline"> = {
  street:          "default",
  railroad:        "secondary",
  utility:         "secondary",
  chance:          "outline",
  community_chest: "outline",
  tax:             "outline",
  go:              "outline",
  jail:            "outline",
  go_to_jail:      "outline",
  free_parking:    "outline",
};

export default function TilesPage() {
  const [tiles, setTiles] = useState<Tile[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<Tile | null>(null);
  const [saving, setSaving] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    getTiles()
      .then(setTiles)
      .catch(() => toast({ title: "Failed to load tiles", variant: "destructive" }))
      .finally(() => setLoading(false));
  }, []);

  const openEdit = (t: Tile) => setEditing({ ...t });

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    try {
      await updateTile(editing.position, {
        name: editing.name,
        description: editing.description,
        icon: editing.icon,
      });
      setTiles((ts) => ts.map((t) => (t.position === editing.position ? editing : t)));
      toast({ title: "Tile saved" });
      setEditing(null);
    } catch {
      toast({ title: "Failed to save", variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6 space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Board Tiles</h1>
        <p className="text-muted-foreground text-sm mt-1">
          All 40 tiles — click a row to edit name, description, and icon
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
                <TableHead className="w-12">#</TableHead>
                <TableHead className="w-8">Color</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Description</TableHead>
                <TableHead className="w-10"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tiles.map((t) => (
                <TableRow key={t.position} className="cursor-pointer" onClick={() => openEdit(t)}>
                  <TableCell className="font-mono text-xs text-muted-foreground">{t.position}</TableCell>
                  <TableCell>
                    {t.color_group ? (
                      <span className={`inline-block h-4 w-4 rounded-sm ${COLOR_MAP[t.color_group] ?? "bg-zinc-400"}`} />
                    ) : (
                      <span className="text-lg leading-none">{t.icon}</span>
                    )}
                  </TableCell>
                  <TableCell className="font-medium">{t.name}</TableCell>
                  <TableCell>
                    <Badge variant={TYPE_VARIANT[t.type] ?? "outline"} className="capitalize text-xs">
                      {t.type.replace("_", " ")}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm max-w-xs truncate">
                    {t.description}
                  </TableCell>
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
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              Edit Tile #{editing?.position} — {editing?.type}
            </DialogTitle>
          </DialogHeader>

          {editing && (
            <div className="space-y-4 py-2">
              <div className="space-y-1.5">
                <Label>Name</Label>
                <Input
                  value={editing.name}
                  onChange={(e) => setEditing({ ...editing, name: e.target.value })}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Description</Label>
                <Textarea
                  value={editing.description}
                  rows={3}
                  onChange={(e) => setEditing({ ...editing, description: e.target.value })}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Icon (emoji or URL)</Label>
                <div className="flex gap-2 items-center">
                  <Input
                    value={editing.icon}
                    onChange={(e) => setEditing({ ...editing, icon: e.target.value })}
                  />
                  <span className="text-2xl">{editing.icon}</span>
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
