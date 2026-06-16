import type { MicrofrontendRoute } from "@moh-sso/microfrontend";
import { announcementsRoute } from "@moh-sso/announcements";
import { auditRoute } from "@moh-sso/audit";
import { clientsRoute } from "@moh-sso/clients";
import { dataValidationRoute } from "@moh-sso/data-validation";
import { dataVisualizerRoute } from "@moh-sso/data-visualizer";
import { documentsRoute } from "@moh-sso/documents";
import { eServicesRoute } from "@moh-sso/e-services";
import { emailRoute } from "@moh-sso/email";
import { issueTrackerRoute } from "@moh-sso/issue-tracker";
import { reportBrowserRoute } from "@moh-sso/report-browser";
import { rbacRoute } from "@moh-sso/rbac";
import { surveillanceRoute } from "@moh-sso/surveillance";
import { usersRoute } from "@moh-sso/users";
import { utilitiesRoute } from "@moh-sso/utilities";

export const registeredMicrofrontends: MicrofrontendRoute[] = [
  announcementsRoute,
  auditRoute,
  clientsRoute,
  dataValidationRoute,
  dataVisualizerRoute,
  documentsRoute,
  eServicesRoute,
  emailRoute,
  issueTrackerRoute,
  reportBrowserRoute,
  rbacRoute,
  surveillanceRoute,
  usersRoute,
  utilitiesRoute,
];

export function registerMicrofrontends() {
  return registeredMicrofrontends;
}
