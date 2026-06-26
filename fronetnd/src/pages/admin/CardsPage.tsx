import { useEffect, useState } from "react";
import { getCards, createCard, updateCard, deleteCard, type Card } from "@/lib/api";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/components/ui/toaster";
import { Loader2, Plus, Pencil, Trash2 } from "lucide-react";

const EFFECT_TYPES = [
  "move_to", "move_relative", "nearest_railroad", "nearest_utility",
  "collect", "pay", "collect_per_player", "pay_per_player",
  "pay_per_building", "collect_per_building", "go_to_jail",
  "get_out_of_jail", "back_to_go",
];

const EFFECT_PAYLOAD_HINT: Record<string, string> = {
  move_to:               '{"position":0,"collect_go":true}',
  move_relative:         '{"steps":-3}',
  nearest_railroad:      '{"double_rent":true}',
  nearest_utility:       '{"dice_mult":10}',
  collect:               '{"amount":50}',
  pay:                   '{"amount":50}',
  collect_per_player:    '{"amount":50}',
  pay_per_player:        '{"amount":50}',
  pay_per_building:      '{"house":25,"hotel":100}',
  collect_per_building:  '{"house":25,"hotel":100}',
  go_to_jail:            '{}',
  get_out_of_jail:       '{}',
  back_to_go:            '{}',
};

type Deck = "chance" | "community_chest";

const blank = (deck: Deck): Omit<Card, "id"> => ({
  deck,
  name: "",
  description: "",
  effect_type: "collect",
  effect_payload: { amount: 0 },
  is_active: true,
  sort_order: 0,
});

export default function CardsPage() {
  const [cards, setCards] = useState<Card[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<Card | null>(null);
  const [creating, setCreating] = useState<Omit<Card, "id"> | null>(null);
  const [payloadText, setPayloadText] = useState("{}");
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    getCards()
      .then(setCards)
      .catch(() => toast({ title: "Failed to load cards", variant: "destructive" }))
      .finally(() => setLoading(false));
  }, []);

  const openCreate = (deck: Deck) => {
    const b = blank(deck);
    setCreating(b);
    setPayloadText("{}");
  };

  const openEdit = (c: Card) => {
    setEditing({ ...c });
    setPayloadText(JSON.stringify(c.effect_payload, null, 2));
  };

  const parsePayload = (): Record<string, unknown> | null => {
    try { return JSON.parse(payloadText); }
    catch { toast({ title: "Invalid JSON payload", variant: "destructive" }); return null; }
  };

  const saveEdit = async () => {
    if (!editing) return;
    const payload = parsePayload();
    if (!payload) return;
    setSaving(true);
    try {
      await updateCard(editing.id, { ...editing, effect_payload: payload });
      setCards((cs) => cs.map((c) => c.id === editing.id ? { ...editing, effect_payload: payload } : c));
      toast({ title: "Card updated" });
      setEditing(null);
    } catch {
      toast({ title: "Failed to save", variant: "destructive" });
    } finally { setSaving(false); }
  };

  const saveCreate = async () => {
    if (!creating) return;
    const payload = parsePayload();
    if (!payload) return;
    setSaving(true);
    try {
      const created = await createCard({ ...creating, effect_payload: payload });
      setCards((cs) => [...cs, created]);
      toast({ title: "Card created" });
      setCreating(null);
    } catch {
      toast({ title: "Failed to create", variant: "destructive" });
    } finally { setSaving(false); }
  };

  const remove = async (id: string) => {
    setDeleting(id);
    try {
      await deleteCard(id);
      setCards((cs) => cs.filter((c) => c.id !== id));
      toast({ title: "Card deleted" });
    } catch {
      toast({ title: "Failed to delete", variant: "destructive" });
    } finally { setDeleting(null); }
  };

  const toggleActive = async (card: Card) => {
    try {
      await updateCard(card.id, { is_active: !card.is_active });
      setCards((cs) => cs.map((c) => c.id === card.id ? { ...c, is_active: !c.is_active } : c));
    } catch {
      toast({ title: "Failed to update", variant: "destructive" });
    }
  };

  const CardTable = ({ deck }: { deck: Deck }) => {
    const list = cards.filter((c) => c.deck === deck);
    return (
      <div className="space-y-3">
        <div className="flex justify-end">
          <Button size="sm" onClick={() => openCreate(deck)}>
            <Plus className="h-4 w-4" /> Add Card
          </Button>
        </div>
        <div className="rounded-xl border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Effect</TableHead>
                <TableHead className="w-20 text-center">Active</TableHead>
                <TableHead className="w-20"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {list.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground py-10">No cards</TableCell>
                </TableRow>
              )}
              {list.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="font-medium">{c.name}</TableCell>
                  <TableCell className="text-muted-foreground text-sm max-w-xs truncate">{c.description}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-xs font-mono">{c.effect_type}</Badge>
                  </TableCell>
                  <TableCell className="text-center">
                    <Switch checked={c.is_active} onCheckedChange={() => toggleActive(c)} />
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1 justify-end">
                      <Button size="icon" variant="ghost" className="h-8 w-8" onClick={() => openEdit(c)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        size="icon" variant="ghost"
                        className="h-8 w-8 text-destructive hover:text-destructive"
                        onClick={() => remove(c.id)}
                        disabled={deleting === c.id}
                      >
                        {deleting === c.id
                          ? <Loader2 className="h-3.5 w-3.5 animate-spin" />
                          : <Trash2 className="h-3.5 w-3.5" />}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    );
  };

  const CardForm = ({ data, onChange }: {
    data: Partial<Card>;
    onChange: (k: string, v: unknown) => void;
  }) => (
    <div className="space-y-4 py-2">
      <div className="space-y-1.5">
        <Label>Name</Label>
        <Input value={data.name ?? ""} onChange={(e) => onChange("name", e.target.value)} />
      </div>
      <div className="space-y-1.5">
        <Label>Description (shown on card face)</Label>
        <Textarea rows={2} value={data.description ?? ""} onChange={(e) => onChange("description", e.target.value)} />
      </div>
      <div className="space-y-1.5">
        <Label>Effect Type</Label>
        <Select
          value={data.effect_type ?? "collect"}
          onValueChange={(v) => {
            onChange("effect_type", v);
            setPayloadText(EFFECT_PAYLOAD_HINT[v] ?? "{}");
          }}
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            {EFFECT_TYPES.map((t) => (
              <SelectItem key={t} value={t}>{t}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-1.5">
        <Label>Effect Payload (JSON)</Label>
        <Textarea
          rows={3}
          value={payloadText}
          onChange={(e) => setPayloadText(e.target.value)}
          className="font-mono text-xs"
          placeholder={EFFECT_PAYLOAD_HINT[data.effect_type ?? "collect"]}
        />
        <p className="text-xs text-muted-foreground">
          Hint: {EFFECT_PAYLOAD_HINT[data.effect_type ?? "collect"]}
        </p>
      </div>
      <div className="flex items-center gap-3">
        <Switch
          checked={data.is_active ?? true}
          onCheckedChange={(v) => onChange("is_active", v)}
        />
        <Label>Active (included in shuffled deck)</Label>
      </div>
    </div>
  );

  return (
    <div className="p-6 space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Cards</h1>
        <p className="text-muted-foreground text-sm mt-1">
          Chance and Community Chest decks — toggle active, edit effects, add custom cards
        </p>
      </div>

      {loading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <Tabs defaultValue="chance">
          <TabsList>
            <TabsTrigger value="chance">
              Chance ({cards.filter((c) => c.deck === "chance").length})
            </TabsTrigger>
            <TabsTrigger value="community_chest">
              Community Chest ({cards.filter((c) => c.deck === "community_chest").length})
            </TabsTrigger>
          </TabsList>
          <TabsContent value="chance"><CardTable deck="chance" /></TabsContent>
          <TabsContent value="community_chest"><CardTable deck="community_chest" /></TabsContent>
        </Tabs>
      )}

      {/* Edit dialog */}
      <Dialog open={!!editing} onOpenChange={(o) => !o && setEditing(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>Edit Card</DialogTitle></DialogHeader>
          {editing && (
            <CardForm
              data={editing}
              onChange={(k, v) => setEditing((e) => e ? { ...e, [k]: v } : e)}
            />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditing(null)}>Cancel</Button>
            <Button onClick={saveEdit} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 animate-spin" />} Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create dialog */}
      <Dialog open={!!creating} onOpenChange={(o) => !o && setCreating(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>
              New {creating?.deck === "chance" ? "Chance" : "Community Chest"} Card
            </DialogTitle>
          </DialogHeader>
          {creating && (
            <CardForm
              data={creating}
              onChange={(k, v) => setCreating((c) => c ? { ...c, [k]: v } : c)}
            />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreating(null)}>Cancel</Button>
            <Button onClick={saveCreate} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 animate-spin" />} Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
