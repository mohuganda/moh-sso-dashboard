# Non-React Microfrontends

Future apps can use Vue, Angular, Svelte, or plain JavaScript as long as they expose the same lifecycle contract.

## Required Exports

```ts
export async function bootstrap(props) {}
export async function mount(props) {}
export async function unmount(props) {}
```

Each app should accept `MicrofrontendRuntimeProps` from `@moh-sso/microfrontend`.

## Framework Adapters

- Vue: `single-spa-vue`
- Angular: `single-spa-angular`
- Svelte: `single-spa-svelte`
- Plain JavaScript: export lifecycle functions directly

Do not add these frameworks until an actual app needs them.

## Rules

- Do not import shell internals.
- Use shared packages for API, auth contracts, types, and UI contracts.
- Scope CSS to avoid leaking styles.
- Communicate with the shell through props or `eventBus`.
