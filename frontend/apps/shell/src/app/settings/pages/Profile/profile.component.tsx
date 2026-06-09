import { useEffect, useMemo, useRef, useState } from "react";
import {
  Button,
  Form,
  InlineLoading,
  InlineNotification,
  SkeletonText,
  Stack,
  Tag,
  TextInput,
  Tile,
} from "@carbon/react";
import { CheckmarkFilled, Save, Security, UserAvatar, UserRole, Time } from "@carbon/react/icons";

import { useMeQuery, useUpdateProfileMutation } from "@moh-sso/api";
import UserSessionsTable from "@/app/settings/sessions/UserSessionsTable";

import "./profile.scss";

type ProfileFormState = {
  firstName: string;
  lastName: string;
};

function formatDate(date?: string | null): string {
  if (!date) {
    return "—";
  }

  const parsedDate = new Date(date);

  if (Number.isNaN(parsedDate.getTime())) {
    return "—";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(parsedDate);
}

function getInitials(user?: {
  firstName?: string | null;
  lastName?: string | null;
  fullName?: string | null;
  username?: string | null;
  email?: string | null;
}): string {
  const firstName = user?.firstName?.trim();
  const lastName = user?.lastName?.trim();

  if (firstName && lastName) {
    return `${firstName[0]}${lastName[0]}`.toUpperCase();
  }

  const fullNameParts = user?.fullName?.trim().split(/\s+/).filter(Boolean);

  if (fullNameParts && fullNameParts.length >= 2) {
    return `${fullNameParts[0][0]}${fullNameParts[1][0]}`.toUpperCase();
  }

  const fallback = user?.username || user?.email || "User";

  return fallback.slice(0, 2).toUpperCase();
}

function getDisplayName(user?: {
  fullName?: string | null;
  username?: string | null;
  email?: string | null;
}): string {
  return user?.fullName || user?.username || user?.email || "My Profile";
}

function getRoleCount(user?: {
  realmRoles?: string[] | null;
  clientRoles?: Record<string, string[]> | null;
}): number {
  const realmRoleCount = user?.realmRoles?.length ?? 0;

  const clientRoleCount = Object.values(user?.clientRoles ?? {}).reduce(
    (total, roles) => total + roles.length,
    0,
  );

  return realmRoleCount + clientRoleCount;
}

function ProfileLoadingState() {
  return (
    <div className="profile-page">
      <Tile className="profile-card profile-header-card">
        <div className="profile-header">
          <div className="profile-avatar profile-avatar--loading" />
          <div className="profile-header__content">
            <SkeletonText width="30%" />
            <SkeletonText width="45%" />
            <SkeletonText width="25%" />
          </div>
        </div>
      </Tile>

      <div className="profile-summary-grid">
        {Array.from({ length: 3 }).map((_, index) => (
          <Tile className="profile-card" key={index}>
            <SkeletonText width="40%" />
            <SkeletonText paragraph lineCount={2} />
          </Tile>
        ))}
      </div>
    </div>
  );
}

export default function MyProfilePage() {
  const sessionsRef = useRef<HTMLDivElement | null>(null);

  const { data: user, isLoading, isFetching, isError, refetch } = useMeQuery();

  const [updateProfile, { isLoading: isSaving }] = useUpdateProfileMutation();

  const [form, setForm] = useState<ProfileFormState>({
    firstName: "",
    lastName: "",
  });

  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    if (!user) {
      return;
    }

    setForm({
      firstName: user.firstName || "",
      lastName: user.lastName || "",
    });
  }, [user]);

  const initialForm = useMemo<ProfileFormState>(
    () => ({
      firstName: user?.firstName || "",
      lastName: user?.lastName || "",
    }),
    [user],
  );

  const isDirty = useMemo(() => {
    return (
      form.firstName.trim() !== initialForm.firstName.trim() ||
      form.lastName.trim() !== initialForm.lastName.trim()
    );
  }, [form, initialForm]);

  const initials = useMemo(() => getInitials(user), [user]);

  const totalRoles = useMemo(() => getRoleCount(user), [user]);

  const securityTags = useMemo(() => {
    const tags: Array<{ label: string; type: "red" | "purple" | "green" | "cool-gray" }> = [];

    if (!user?.enabled) {
      tags.push({ label: "Account Disabled", type: "red" });
    }

    if (!user?.emailVerified) {
      tags.push({ label: "Email Unverified", type: "red" });
    }

    if (user?.requirePwdChange) {
      tags.push({ label: "Password Change Due", type: "purple" });
    }

    if (user?.enabled && user?.emailVerified && !user?.requirePwdChange) {
      tags.push({ label: "Secure", type: "green" });
    }

    if (tags.length === 0) {
      tags.push({ label: "No security status available", type: "cool-gray" });
    }

    return tags;
  }, [user]);

  const handleFormChange = (field: keyof ProfileFormState, value: string) => {
    setSuccessMessage("");
    setErrorMessage("");

    setForm((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const handleReset = () => {
    setSuccessMessage("");
    setErrorMessage("");
    setForm(initialForm);
  };

  const handleManageSessions = () => {
    sessionsRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    setSuccessMessage("");
    setErrorMessage("");

    try {
      await updateProfile({
        firstName: form.firstName.trim(),
        lastName: form.lastName.trim(),
      }).unwrap();

      setSuccessMessage("Your profile has been updated successfully.");
    } catch (error) {
      console.error(error);
      setErrorMessage("Unable to update your profile. Please try again.");
    }
  };

  if (isLoading) {
    return <ProfileLoadingState />;
  }

  if (isError || !user) {
    return (
      <div className="profile-page">
        <InlineNotification
          kind="error"
          title="Failed to load profile"
          subtitle="We could not load your profile details at the moment."
          lowContrast
        />
      </div>
    );
  }

  return (
    <div className="profile-page">
      {isFetching && (
        <div className="profile-refreshing">
          <InlineLoading description="Refreshing profile..." />
        </div>
      )}

      {successMessage && (
        <InlineNotification
          kind="success"
          title="Profile updated"
          subtitle={successMessage}
          lowContrast
          onClose={() => setSuccessMessage("")}
        />
      )}

      {errorMessage && (
        <InlineNotification
          kind="error"
          title="Update failed"
          subtitle={errorMessage}
          lowContrast
          onClose={() => setErrorMessage("")}
        />
      )}

      {/* PROFILE HEADER */}
      <Tile className="profile-card profile-header-card">
        <div className="profile-header">
          <div className="profile-avatar" aria-hidden="true">
            {initials}
          </div>

          <div className="profile-header__content">
            <div className="profile-header__title-row">
              <h3 className="profile-header__title">{getDisplayName(user)}</h3>

              {user.enabled ? (
                <Tag size="sm" type="green">
                  Active
                </Tag>
              ) : (
                <Tag size="sm" type="red">
                  Disabled
                </Tag>
              )}
            </div>

            <p className="profile-header__email">{user.email || "No email available"}</p>

            <div className="profile-header__meta">
              <Time size={14} />
              <span>Last login: {formatDate(user.lastLoginAt)}</span>
            </div>
          </div>
        </div>
      </Tile>

      {/* DASHBOARD SUMMARY */}
      <div className="profile-summary-grid">
        <Tile className="profile-card profile-summary-card">
          <div className="profile-card__heading">
            <UserAvatar size={18} />
            <h4>Activity Summary</h4>
          </div>

          <dl className="profile-stat-list">
            <div>
              <dt>Created</dt>
              <dd>{formatDate(user.createdAt)}</dd>
            </div>

            <div>
              <dt>Total Roles</dt>
              <dd>
                <Tag size="sm" type="cool-gray">
                  {totalRoles}
                </Tag>
              </dd>
            </div>
          </dl>
        </Tile>

        <Tile className="profile-card profile-summary-card">
          <div className="profile-card__heading">
            <Security size={18} />
            <h4>Security Alerts</h4>
          </div>

          <div className="profile-tag-list">
            {securityTags.map((tag) => (
              <Tag key={tag.label} size="sm" type={tag.type}>
                {tag.label}
              </Tag>
            ))}
          </div>
        </Tile>

        <Tile className="profile-card profile-summary-card">
          <div className="profile-card__heading">
            <CheckmarkFilled size={18} />
            <h4>Session</h4>
          </div>

          <Stack gap={3}>
            <p className="profile-card__text">Active: Current browser</p>

            <Button size="sm" kind="ghost" onClick={handleManageSessions}>
              Manage Sessions
            </Button>
          </Stack>
        </Tile>
      </div>

      {/* ACCOUNT + PROFILE */}
      <div className="profile-two-column">
        <Tile className="profile-card">
          <div className="profile-card__heading profile-card__heading--spaced">
            <UserAvatar size={18} />
            <h4>Account Details</h4>
          </div>

          <Stack gap={4}>
            <TextInput
              id="profile-username"
              labelText="Username"
              value={user.username || ""}
              readOnly
              size="sm"
            />

            <TextInput
              id="profile-email"
              labelText="Email"
              value={user.email || ""}
              readOnly
              size="sm"
            />

            <TextInput
              id="profile-full-name"
              labelText="Full Name"
              value={user.fullName || ""}
              readOnly
              size="sm"
            />
          </Stack>
        </Tile>

        <Tile className="profile-card">
          <div className="profile-card__heading profile-card__heading--spaced">
            <UserAvatar size={18} />
            <h4>Update Profile</h4>
          </div>

          <Form onSubmit={handleSubmit}>
            <Stack gap={4}>
              <TextInput
                id="profile-first-name"
                labelText="First Name"
                value={form.firstName}
                size="sm"
                disabled={isSaving}
                onChange={(event) => handleFormChange("firstName", event.target.value)}
              />

              <TextInput
                id="profile-last-name"
                labelText="Last Name"
                value={form.lastName}
                size="sm"
                disabled={isSaving}
                onChange={(event) => handleFormChange("lastName", event.target.value)}
              />

              <div className="profile-form-actions">
                <Button
                  size="sm"
                  type="submit"
                  disabled={isSaving || !isDirty}
                  renderIcon={isSaving ? undefined : Save}
                >
                  {isSaving ? "Saving..." : "Save Changes"}
                </Button>

                <Button
                  size="sm"
                  kind="ghost"
                  type="button"
                  disabled={isSaving || !isDirty}
                  onClick={handleReset}
                >
                  Reset
                </Button>
              </div>
            </Stack>
          </Form>
        </Tile>
      </div>

      {/* ROLES */}
      <Tile className="profile-card">
        <div className="profile-card__heading profile-card__heading--spaced">
          <UserRole size={18} />
          <h4>Roles & Access</h4>
        </div>

        <div className="profile-roles-grid">
          <section>
            <p className="profile-section-label">Realm Roles</p>

            {user.realmRoles && user.realmRoles.length > 0 ? (
              <div className="profile-tag-list">
                {user.realmRoles.map((role) => (
                  <Tag key={role} size="sm">
                    {role}
                  </Tag>
                ))}
              </div>
            ) : (
              <p className="profile-empty-text">No realm roles assigned.</p>
            )}
          </section>

          <section>
            <p className="profile-section-label">Client Roles</p>

            {user.clientRoles && Object.keys(user.clientRoles).length > 0 ? (
              <div className="profile-client-roles">
                {Object.entries(user.clientRoles).map(([client, roles]) => (
                  <div className="profile-client-role-group" key={client}>
                    <p className="profile-client-role-group__title">{client}</p>

                    {roles.length > 0 ? (
                      <div className="profile-tag-list">
                        {roles.map((role) => (
                          <Tag key={`${client}-${role}`} type="cyan" size="sm">
                            {role}
                          </Tag>
                        ))}
                      </div>
                    ) : (
                      <p className="profile-empty-text">No roles assigned.</p>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <p className="profile-empty-text">No client roles assigned.</p>
            )}
          </section>
        </div>
      </Tile>

      {/* SESSIONS */}
      <Tile className="profile-card" ref={sessionsRef}>
        <div className="profile-card__heading profile-card__heading--spaced">
          <Security size={18} />
          <h4>Active Sessions</h4>
        </div>

        <UserSessionsTable />
      </Tile>
    </div>
  );
}
