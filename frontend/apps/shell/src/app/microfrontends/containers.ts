function containerId(appName: string) {
  return `single-spa-application-${appName.replace(/[^a-zA-Z0-9_-]/g, "-")}`;
}

function getOrCreateRoot() {
  const existingRoot = document.getElementById("microfrontend-orchestrated-root");
  if (existingRoot) {
    return existingRoot;
  }

  const root = document.createElement("div");
  root.id = "microfrontend-orchestrated-root";
  document.body.appendChild(root);
  return root;
}

export function getMicrofrontendContainer(appName: string) {
  const id = containerId(appName);
  const existingContainer = document.getElementById(id);
  if (existingContainer) {
    return existingContainer;
  }

  const container = document.createElement("div");
  container.id = id;
  container.dataset.microfrontend = appName;
  getOrCreateRoot().appendChild(container);
  return container;
}
