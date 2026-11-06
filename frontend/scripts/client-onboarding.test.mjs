import test from 'node:test';
import assert from 'node:assert/strict';
import { emptyDraft, parseOnboardingDraft, exportOnboardingDraft } from '../apps/clients/src/components/onboarding-draft.ts';

test('promotion keeps source environment until the admin reviews the target', () => {
  const draft = { ...emptyDraft('development'), clientId: 'timesheets', name: 'Timesheets', owner: 'Workforce', supportEmail: 'support@example.invalid' };
  assert.deepEqual(parseOnboardingDraft(exportOnboardingDraft(draft)), draft);
  assert.equal(parseOnboardingDraft(exportOnboardingDraft(draft)).environment, 'development');
});
test('rejects credentials, full realm exports, prototype keys and malformed documents', () => {
  const draft = emptyDraft('staging');
  for (const value of [{ ...draft, secret: 'fixture' }, { ...draft, credentials: [] }, { ...draft, users: [] }, { ...draft, version: 2 }, { ...draft, redirectUris: [42] }, { ...draft, launcher: 'true' }, { ...draft, constructor: 'unexpected' }]) {
    assert.throws(() => parseOnboardingDraft(JSON.stringify(value)));
  }
});
