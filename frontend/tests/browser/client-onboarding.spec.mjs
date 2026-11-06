import { test, expect } from '@playwright/test';
import { URL } from 'node:url';

const admin = { id: '00000000-0000-0000-0000-000000000001', username: 'admin', email: 'admin@example.invalid', isAdmin: true, enabled: true, realmRoles: ['admin'], clientRoles: {}, permissions: ['portal:access', 'clients:read', 'clients:write'], systems: [], accessibleSystems: [] };

for (const viewport of [{ width: 1280, height: 900 }, { width: 390, height: 844 }]) {
test(`admin reviews and provisions a disabled client at ${viewport.width}px`, async ({ page }) => {
  await page.setViewportSize(viewport);
  const writes = [];
  const failures = [];
  page.on('pageerror', error => failures.push(error.message));
  await page.route('http://127.0.0.1:5417/ssobackend/**', async route => {
    const path = new URL(route.request().url()).pathname;
    let data = [];
    if (path.endsWith('/auth/me')) data = { user: admin };
    else if (path.includes('/onboarding/context')) data = { environment: 'staging', realm: 'moh-staging', issuer: 'https://auth.example.invalid/realms/moh-staging', profiles: [{ id: 'public-web', name: 'Public web application (PKCE)' }] };
    else if (path.endsWith('/onboarding/preview')) {
      const { spec } = route.request().postDataJSON();
      data = { spec, operation: 'create', realm: 'moh-staging', issuer: 'https://auth.example.invalid/realms/moh-staging', reviewToken: 'fixture-review', client: { ...spec, enabled: false, publicClient: true, standardFlowEnabled: true, defaultClientScopes: ['profile', 'email'], attributes: { 'pkce.code.challenge.method': 'S256' } } };
    } else if (path.endsWith('/onboarding/apply')) {
      const body = route.request().postDataJSON(); writes.push(body);
      data = { client: { ...body.spec, id: 'fixture', enabled: false }, issuer: 'https://auth.example.invalid/realms/moh-staging', message: 'Provisioned and verified. Client is disabled.' };
    } else if (path.includes('unread')) data = { count: 0 };
    await route.fulfill({ json: { success: true, data } });
  });
  await page.goto('/portal/admin/clients');
  await page.getByRole('button', { name: 'Onboard application', exact: true }).click();
  await page.getByLabel('Client ID', { exact: true }).fill('workforce-app');
  await page.getByLabel('Application name', { exact: true }).fill('Workforce');
  await page.getByLabel('Application owner or team').fill('Workforce team');
  await page.getByLabel('Support email', { exact: true }).fill('support@example.invalid');
  await page.getByRole('button', { name: 'Continue', exact: true }).click();
  await page.getByLabel('Application URL', { exact: true }).fill('https://workforce.example.invalid');
  await page.getByLabel('Sign-in callback URLs').fill('https://workforce.example.invalid/callback');
  await page.getByLabel('Post-logout URLs').fill('https://workforce.example.invalid/login');
  await page.getByLabel('Allowed browser origins').fill('https://workforce.example.invalid');
  await page.getByRole('button', { name: 'Validate & preview' }).click();
  const apply = page.getByRole('button', { name: 'Provision disabled client' });
  await expect(apply).toBeDisabled();
  await page.getByText('I reviewed these URLs and settings for staging.', { exact: true }).click();
  await apply.click();
  await expect(page.getByRole('heading', { name: 'Application configured' })).toBeVisible();
  expect(writes).toHaveLength(1);
  expect(writes[0].spec.environment).toBe('staging');
  expect(writes[0].reviewToken).toBe('fixture-review');
  expect(writes[0]).not.toHaveProperty('secret');
  expect(failures).toEqual([]);
});

}
