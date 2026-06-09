// src/config/appMeta.ts
export const APP_ENV = import.meta.env.VITE_APP_ENV || import.meta.env.MODE;

export const APP_VERSION = import.meta.env.VITE_APP_VERSION || "dev";

export const BUILD_TIME = import.meta.env.VITE_BUILD_TIME || "";

