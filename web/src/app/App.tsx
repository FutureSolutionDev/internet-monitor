import { LockKeyhole, RefreshCw, ServerOff, Users } from "lucide-react";
import { useEffect } from "react";
import { Shell } from "../components/Shell";
import { Toaster } from "../components/Toast";
import { Button, Card } from "../components/ui";
import { useI18n } from "../lib/i18n";
import { live, useLive } from "../lib/live";
import { useLocation } from "../lib/router";
import { History } from "../screens/History";
import { Overview } from "../screens/Overview";
import { Report } from "../screens/Report";
import { Speed } from "../screens/Speed";
import { Settings } from "../screens/settings/Settings";

function Gate({ kind }: { kind: "unauthorized" | "busy" | "offline" }) {
  const { t } = useI18n();
  const Icon = kind === "unauthorized" ? LockKeyhole : kind === "busy" ? Users : ServerOff;
  const title =
    kind === "unauthorized"
      ? "gate_unauth_title"
      : kind === "busy"
        ? "gate_busy_title"
        : "gate_offline_title";
  const body =
    kind === "unauthorized" ? "gate_unauth_body" : kind === "busy" ? "gate_busy_body" : "gate_offline_body";
  return (
    <main className="gate">
      <Card>
        <span className="icon-well" style={{ width: 56, height: 56 }}>
          <Icon size={26} aria-hidden />
        </span>
        <h1>{t(title)}</h1>
        <p className="secondary">{t(body)}</p>
        {kind !== "unauthorized" ? (
          <Button variant="edge" icon={RefreshCw} onClick={() => window.location.reload()}>
            {t("retry")}
          </Button>
        ) : null}
      </Card>
    </main>
  );
}

export function App() {
  const { t } = useI18n();
  const { route, params } = useLocation();
  const connection = useLive((s) => s.connection);
  const hasSnapshot = useLive((s) => s.snapshot !== null);

  useEffect(() => {
    live.start();
    return () => live.stop();
  }, []);

  useEffect(() => {
    if (route !== "report") document.title = t("app_name");
  }, [route, t]);

  let body: React.ReactNode;
  if (connection === "unauthorized") body = <Gate kind="unauthorized" />;
  else if (connection === "busy" && !hasSnapshot) body = <Gate kind="busy" />;
  else if (connection === "reconnecting" && !hasSnapshot) body = <Gate kind="offline" />;
  else if (route === "report") body = <Report month={params.get("month")} />;
  else {
    body = (
      <Shell route={route}>
        {route === "history" ? <History /> : null}
        {route === "speed" ? <Speed /> : null}
        {route === "settings" ? <Settings /> : null}
        {route === "overview" ? <Overview /> : null}
      </Shell>
    );
  }

  return (
    <>
      {body}
      <Toaster closeLabel={t("close")} />
    </>
  );
}
