import { BrowserRouter as Router, Routes, Route, Navigate, Outlet } from "react-router-dom";

import { AdminRoute } from "./components/AdminRoute";
import { ModalProvider } from "./components/modal/modal.context";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { UserRoute } from "./components/UserRoute";
import AdminLayout from "./layout/admin/AdminLayout";
import PublicLayout from "./layout/public/PublicLayout";
import UserLayout from "./layout/user/UserLayout";

/* -----------------------------
 * Public pages
 * ----------------------------- */
import AuditLogsPage from "./pages/admin/audit/audit_component";
import ClientsPage from "./pages/admin/clients/client.component";
import HomePage from "./pages/admin/home/home.component";
import UsersPage from "./pages/admin/user/user.component";
import DataVisualizer from "./pages/public/datavisualizer/data-visualizer";
import NewsFeedPage from "./pages/public/newsfeed/news_feed.component";

/* -----------------------------
 * User / client pages
 * ----------------------------- */

/* -----------------------------
 * Admin pages
 * ----------------------------- */
import AbsenceRequests from "./pages/user/utilities/absence-requests.component";
import Elearning from "./pages/user/utilities/elearning.component";
import LeavePlan from "./pages/user/utilities/leave-plan.component";
import MyAbsenceDashboard from "./pages/user/utilities/my-absence-dashboard.component";
import MyTimeSheet from "./pages/user/utilities/my-timesheet.component";
import FileUpload from "./pages/public/fileupload/file-upload.tsx";
import DocumentPage from "./pages/user/e-services/documents/documents.component";
import DocumentDetailsPage from "./pages/user/e-services/documents/document-details.component";
import MyProfilePage from "./pages/user/settings/Profile/profile.component.tsx";
import SecurityPage from "./pages/user/settings/security/security.component.tsx";
import ActiveSessionsPage from "./pages/user/settings/sessions/active-sesssions.component.tsx";
import { AnnouncementsPage } from "./pages/admin/announcements/announcements.components.tsx";
import SurveillancePage from "./pages/user/surveillence/surveillance.component.tsx";
import DiseaseDetailsPage from "./pages/user/surveillence/surveillance-details/surveillance-details.component.tsx";

function App() {
  return (
    <ModalProvider>
      <Router basename={"/portal"}>
        <Routes>
          {/* ================================================== */}
          {/* PUBLIC */}
          {/* ================================================== */}
          <Route element={<PublicLayout />}>
            <Route index element={<NewsFeedPage />} />
            <Route path="/" element={<NewsFeedPage />} />
          </Route>

          {/* ================================================== */}
          {/* USER (authenticated) */}
          {/* ================================================== */}
          <Route
            path="/apps"
            element={
              <ProtectedRoute>
                <UserLayout />
              </ProtectedRoute>
            }
          >
            {/* /apps -> choose a sensible default */}
            <Route index element={<Navigate to="news" replace />} />

            {/* simple: keep news directly under /apps */}
            <Route
              path="news"
              element={
                <UserRoute>
                  <NewsFeedPage />
                </UserRoute>
              }
            />

            {/* --------------------------
             * DWH client
             * -------------------------- */}
            <Route path="dwh">
              <Route index element={<Navigate to="data-visualizer" replace />} />
              <Route
                path="data-visualizer"
                element={
                  <UserRoute>
                    <DataVisualizer />
                  </UserRoute>
                }
              />
              <Route
                path="dashboards"
                element={
                  <UserRoute>
                    <div>Dashboards</div>
                  </UserRoute>
                }
              />
              <Route
                path="reports"
                element={
                  <UserRoute>
                    <div>Reports</div>
                  </UserRoute>
                }
              />
              <Route
                path="exports"
                element={
                  <UserRoute>
                    <div>Data Exports</div>
                  </UserRoute>
                }
              />
              <Route
                path="filesvr"
                element={
                  <UserRoute>
                    <FileUpload />
                  </UserRoute>
                }
              />

              <Route
                path="surveillance"
                element={
                  <UserRoute>
                    <Outlet />
                  </UserRoute>
                }
              >
                <Route index element={<SurveillancePage />} />
                <Route path=":diseaseName" element={<DiseaseDetailsPage />} />
              </Route>
            </Route>

            {/* --------------------------
             * eServices client
             * -------------------------- */}
            <Route path="eservices">
              <Route index element={<Navigate to="ihris" replace />} />

              <Route
                path="ihris"
                element={
                  <UserRoute>
                    <div>iHRIS</div>
                  </UserRoute>
                }
              />
              <Route
                path="meeting-manager"
                element={
                  <UserRoute>
                    <div>Meeting Manager</div>
                  </UserRoute>
                }
              />
              <Route
                path="action-tracker"
                element={
                  <UserRoute>
                    <div>Action Tracker</div>
                  </UserRoute>
                }
              />
              <Route
                path="clinician-outputs"
                element={
                  <UserRoute>
                    <div>Clinician Outputs</div>
                  </UserRoute>
                }
              />
              <Route
                path="leave-absence"
                element={
                  <UserRoute>
                    <div>Leave & Absence</div>
                  </UserRoute>
                }
              />
              <Route
                path="workplans"
                element={
                  <UserRoute>
                    <div>Workplans</div>
                  </UserRoute>
                }
              />
              <Route
                path="budget-tracker"
                element={
                  <UserRoute>
                    <div>Budget Tracker</div>
                  </UserRoute>
                }
              />
              <Route
                path="activity-reporting"
                element={
                  <UserRoute>
                    <div>Activity Reporting</div>
                  </UserRoute>
                }
              />
              <Route
                path="partner-management"
                element={
                  <UserRoute>
                    <div>Partner Management</div>
                  </UserRoute>
                }
              />
              <Route
                path="observatory-uploads"
                element={
                  <UserRoute>
                    <div>Observatory Uploads</div>
                  </UserRoute>
                }
              />
            </Route>

            {/* --------------------------
             * Research & Studies client
             * -------------------------- */}
            <Route path="research-studies">
              <Route index element={<Navigate to="studies" replace />} />

              <Route
                path="studies"
                element={
                  <UserRoute>
                    <div>Studies</div>
                  </UserRoute>
                }
              />
              <Route
                path="datasets"
                element={
                  <UserRoute>
                    <div>Datasets</div>
                  </UserRoute>
                }
              />
              <Route
                path="ethics"
                element={
                  <UserRoute>
                    <div>Ethics & Approvals</div>
                  </UserRoute>
                }
              />
              <Route
                path="publications"
                element={
                  <UserRoute>
                    <div>Publications</div>
                  </UserRoute>
                }
              />
            </Route>

            {/* --------------------------
             * Case Registers client
             * -------------------------- */}
            <Route path="case-registers">
              <Route index element={<Navigate to="external-referrals" replace />} />

              <Route
                path="external-referrals"
                element={
                  <UserRoute>
                    <div>External Referrals</div>
                  </UserRoute>
                }
              />
              <Route
                path="disease-registers"
                element={
                  <UserRoute>
                    <div>Disease Registers</div>
                  </UserRoute>
                }
              />
            </Route>

            {/* --------------------------
             * Outbreak Management client
             * -------------------------- */}
            <Route path="outbreak-management">
              <Route index element={<Navigate to="signals-alerts" replace />} />

              <Route
                path="signals-alerts"
                element={
                  <UserRoute>
                    <div>Signals & Alerts</div>
                  </UserRoute>
                }
              />
              <Route
                path="poe-management"
                element={
                  <UserRoute>
                    <div>PoE Management</div>
                  </UserRoute>
                }
              />
              <Route
                path="case-management"
                element={
                  <UserRoute>
                    <div>Case Management</div>
                  </UserRoute>
                }
              />
            </Route>

            {/* --------------------------
             * Reference Registers client
             * -------------------------- */}
            <Route path="reference-registers">
              <Route index element={<Navigate to="facility-register" replace />} />

              <Route
                path="facility-register"
                element={
                  <UserRoute>
                    <div>Facility Register</div>
                  </UserRoute>
                }
              />

              <Route path="terminology">
                <Route index element={<Navigate to="test-menu" replace />} />
                <Route
                  path="test-menu"
                  element={
                    <UserRoute>
                      <div>Test Menu</div>
                    </UserRoute>
                  }
                />
                <Route
                  path="pharmaceuticals"
                  element={
                    <UserRoute>
                      <div>Pharmaceuticals</div>
                    </UserRoute>
                  }
                />
                <Route
                  path="procedures"
                  element={
                    <UserRoute>
                      <div>Procedures</div>
                    </UserRoute>
                  }
                />
                <Route
                  path="equipment"
                  element={
                    <UserRoute>
                      <div>Equipment</div>
                    </UserRoute>
                  }
                />
              </Route>
            </Route>

            {/* --------------------------
             * Utilities client
             * -------------------------- */}
            <Route path="utilities">
              <Route index element={<Navigate to="self-service/timesheet" replace />} />

              <Route path="self-service">
                <Route
                  path="timesheet"
                  element={
                    <UserRoute>
                      <MyTimeSheet />
                    </UserRoute>
                  }
                />
                <Route
                  path="elearning"
                  element={
                    <UserRoute>
                      <Elearning />
                    </UserRoute>
                  }
                />
                <Route
                  path="leave-plan"
                  element={
                    <UserRoute>
                      <LeavePlan />
                    </UserRoute>
                  }
                />
                <Route
                  path="absence-requests"
                  element={
                    <UserRoute>
                      <AbsenceRequests />
                    </UserRoute>
                  }
                />
                <Route
                  path="absence-dashboard"
                  element={
                    <UserRoute>
                      <MyAbsenceDashboard />
                    </UserRoute>
                  }
                />

                <Route path="eservice">
                  <Route
                    path="document-upload"
                    element={
                      <UserRoute>
                        <DocumentPage />
                      </UserRoute>
                    }
                  />

                  <Route
                    path="document-upload/:id"
                    element={
                      <UserRoute>
                        <DocumentDetailsPage />
                      </UserRoute>
                    }
                  />

                  <Route path="service-access" element={<div>Service Access</div>} />
                  <Route path="equipment-request" element={<div>Equipment Request</div>} />
                </Route>
              </Route>
            </Route>

            {/* --------------------------
             * Settings client
             * -------------------------- */}
            <Route path="settings">
              <Route index element={<Navigate to="profile" replace />} />
              <Route path="profile" element={<MyProfilePage />} />
              <Route path="sessions" element={<ActiveSessionsPage />} />
              <Route path="security" element={<SecurityPage />} />
            </Route>
          </Route>

          {/* ================================================== */}
          {/* ADMIN (authenticated + role) */}
          {/* ================================================== */}
          <Route
            path="/admin"
            element={
              <ProtectedRoute>
                <AdminLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="home" replace />} />

            <Route
              path="home"
              element={
                <AdminRoute>
                  <HomePage />
                </AdminRoute>
              }
            />

            <Route
              path="users"
              element={
                <AdminRoute>
                  <UsersPage />
                </AdminRoute>
              }
            />

            <Route
              path="clients"
              element={
                <AdminRoute>
                  <ClientsPage />
                </AdminRoute>
              }
            />

            <Route
              path="audit-logs"
              element={
                <AdminRoute>
                  <AuditLogsPage />
                </AdminRoute>
              }
            />

            <Route
              path="announcements"
              element={
                <AdminRoute>
                  <AnnouncementsPage />
                </AdminRoute>
              }
            />
          </Route>

          {/* ================================================== */}
          {/* FALLBACK */}
          {/* ================================================== */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ModalProvider>
  );
}

export default App;
