import { useEffect, useState } from "react";
import { getConfig, updateConfig } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/components/ui/toaster";
import { Loader2, Save } from "lucide-react";

type Config = Record<string, unknown>;

const numFields = [
  { key: "starting_balance", label: "Starting Balance" },
  { key: "go_amount",        label: "GO Salary" },
  { key: "income_tax",       label: "Income Tax" },
  { key: "luxury_tax",       label: "Luxury Tax" },
  { key: "house_limit",      label: "House Limit" },
  { key: "hotel_limit",      label: "Hotel Limit" },
  { key: "max_players",      label: "Max Players" },
  { key: "turn_timer_secs",  label: "Turn Timer (sec, 0=off)" },
];

const textFields = [
  { key: "currency_name",   label: "Currency Name" },
  { key: "currency_symbol", label: "Currency Symbol" },
  { key: "currency_icon",   label: "Currency Icon (emoji or URL)" },
];

const toggleFields = [
  { key: "free_parking_jackpot", label: "Free Parking Jackpot", desc: "Taxes & fines accumulate on Free Parking" },
  { key: "auction_on_decline",   label: "Auction on Decline",   desc: "Property goes to auction when a player declines to buy" },
  { key: "no_rent_in_jail",      label: "No Rent in Jail",      desc: "Owner cannot collect rent while in jail" },
  { key: "double_salary_on_go",  label: "Double Salary on GO",  desc: "Collect 2× salary for landing exactly on GO" },
  { key: "bankruptcy_to_bank",   label: "Bankruptcy to Bank",   desc: "Off = assets go to the creditor instead" },
];

export default function ConfigPage() {
  const [config, setConfig] = useState<Config | null>(null);
  const [saving, setSaving] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    getConfig().then(setConfig).catch(() => toast({ title: "Failed to load config", variant: "destructive" }));
  }, []);

  const set = (key: string, value: unknown) =>
    setConfig((c) => c ? { ...c, [key]: value } : c);

  const save = async () => {
    if (!config) return;
    setSaving(true);
    try {
      await updateConfig(config);
      toast({ title: "Config saved" });
    } catch {
      toast({ title: "Failed to save", variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  if (!config) return <PageLoader />;

  return (
    <div className="p-6 max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Game Config</h1>
          <p className="text-muted-foreground text-sm mt-1">Global settings applied to every new game</p>
        </div>
        <Button onClick={save} disabled={saving}>
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          Save
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Currency</CardTitle>
          <CardDescription>How money is displayed in the game</CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-3 gap-4">
          {textFields.map(({ key, label }) => (
            <div key={key} className="space-y-1.5">
              <Label>{label}</Label>
              <Input
                value={String(config[key] ?? "")}
                onChange={(e) => set(key, e.target.value)}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Financials &amp; Limits</CardTitle>
          <CardDescription>Starting money, taxes, and bank limits</CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4">
          {numFields.map(({ key, label }) => (
            <div key={key} className="space-y-1.5">
              <Label>{label}</Label>
              <Input
                type="number"
                value={Number(config[key] ?? 0)}
                onChange={(e) => set(key, Number(e.target.value))}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>House Rules</CardTitle>
          <CardDescription>Optional rule variations — off by default</CardDescription>
        </CardHeader>
        <CardContent className="space-y-0">
          {toggleFields.map(({ key, label, desc }, i) => (
            <div key={key}>
              {i > 0 && <Separator className="my-0" />}
              <div className="flex items-center justify-between py-4">
                <div>
                  <p className="text-sm font-medium">{label}</p>
                  <p className="text-xs text-muted-foreground">{desc}</p>
                </div>
                <Switch
                  checked={Boolean(config[key])}
                  onCheckedChange={(v) => set(key, v)}
                />
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center">
      <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
    </div>
  );
}
