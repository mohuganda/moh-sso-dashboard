import { useEffect, useRef } from "react";
import { registerSW } from "virtual:pwa-register";

import { useToast } from "@moh-sso/ui";

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
};

export function PwaLifecycle() {
  const toast = useToast();
  const updateRef = useRef<((reloadPage?: boolean) => Promise<void>) | null>(null);
  const installPromptRef = useRef<BeforeInstallPromptEvent | null>(null);

  useEffect(() => {
    updateRef.current = registerSW({
      immediate: false,
      onNeedRefresh() {
        window.dispatchEvent(new CustomEvent("portal:pwa-update-available"));
        toast.info({
          title: "A new portal version is available",
          subtitle: "Refresh to load the latest application files.",
          timeout: 0,
          actions: [
            {
              label: "Refresh",
              onClick: () => {
                void updateRef.current?.(true);
              },
            },
          ],
        });
      },
      onOfflineReady() {
        window.dispatchEvent(new CustomEvent("portal:pwa-offline-ready"));
        toast.success({
          title: "Portal is ready for offline fallback",
          subtitle: "Live services still require a network connection.",
        });
      },
      onRegisterError(error) {
        console.error("PWA service worker registration failed", error);
      },
    });
  }, [toast]);

  useEffect(() => {
    const handleBeforeInstallPrompt = (event: Event) => {
      event.preventDefault();
      installPromptRef.current = event as BeforeInstallPromptEvent;

      toast.info({
        title: "Install MOH Portal",
        subtitle: "Add the portal to this device for quicker access.",
        timeout: 10000,
        actions: [
          {
            label: "Install",
            onClick: () => {
              const promptEvent = installPromptRef.current;
              installPromptRef.current = null;
              void promptEvent?.prompt();
            },
          },
        ],
      });
    };

    const handleAppInstalled = () => {
      installPromptRef.current = null;
      toast.success("Portal installed", "The portal was added to this device.");
    };

    window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
    window.addEventListener("appinstalled", handleAppInstalled);

    return () => {
      window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
      window.removeEventListener("appinstalled", handleAppInstalled);
    };
  }, [toast]);

  return null;
}
