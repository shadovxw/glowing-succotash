import { useEffect, useState } from "react";
import { getPlayers, type Player } from "@/lib/api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/components/ui/toaster";
import { Loader2, Trophy, Gamepad2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

function winRate(p: Player) {
  if (p.games_played === 0) return 0;
  return Math.round((p.games_won / p.games_played) * 100);
}

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1)   return "just now";
  if (m < 60)  return `${m}m ago`;
  if (m < 1440) return `${Math.floor(m / 60)}h ago`;
  return `${Math.floor(m / 1440)}d ago`;
}

export default function PlayersPage() {
  const [players, setPlayers] = useState<Player[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    getPlayers()
      .then(setPlayers)
      .catch(() => toast({ title: "Failed to load players", variant: "destructive" }))
      .finally(() => setLoading(false));
  }, []);

  const totalGames = players.reduce((s, p) => s + p.games_played, 0);
  const topWins = Math.max(...players.map((p) => p.games_won), 0);

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Players</h1>
        <p className="text-muted-foreground text-sm mt-1">
          Everyone who has connected to any game
        </p>
      </div>

      {!loading && (
        <div className="grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Total Players</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{players.length}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Total Games Played</CardTitle>
            </CardHeader>
            <CardContent className="flex items-center gap-2">
              <Gamepad2 className="h-5 w-5 text-muted-foreground" />
              <p className="text-3xl font-bold">{totalGames}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Most Wins</CardTitle>
            </CardHeader>
            <CardContent className="flex items-center gap-2">
              <Trophy className="h-5 w-5 text-yellow-500" />
              <p className="text-3xl font-bold">{topWins}</p>
            </CardContent>
          </Card>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="rounded-xl border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Player</TableHead>
                <TableHead className="text-right">Games Played</TableHead>
                <TableHead className="text-right">Games Won</TableHead>
                <TableHead className="text-right">Win Rate</TableHead>
                <TableHead className="text-right">Last Seen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {players.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground py-16">
                    No players yet — start a game to see data here
                  </TableCell>
                </TableRow>
              )}
              {players.map((p, i) => (
                <TableRow key={p.bastion_user_id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      {p.avatar ? (
                        <img src={p.avatar} alt="" className="h-8 w-8 rounded-full object-cover" />
                      ) : (
                        <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center text-xs font-medium">
                          {p.display_name[0]?.toUpperCase()}
                        </div>
                      )}
                      <div>
                        <p className="font-medium text-sm">{p.display_name}</p>
                        <p className="text-xs text-muted-foreground truncate max-w-[180px]">
                          {p.bastion_user_id}
                        </p>
                      </div>
                      {i === 0 && p.games_won === topWins && topWins > 0 && (
                        <Trophy className="h-4 w-4 text-yellow-500" />
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right font-mono">{p.games_played}</TableCell>
                  <TableCell className="text-right font-mono">{p.games_won}</TableCell>
                  <TableCell className="text-right">
                    <Badge
                      variant={winRate(p) >= 50 ? "default" : "secondary"}
                      className="font-mono"
                    >
                      {winRate(p)}%
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground text-sm">
                    {timeAgo(p.last_seen)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
