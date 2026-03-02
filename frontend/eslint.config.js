import js from "@eslint/js";
import globals from "globals";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import importPlugin from "eslint-plugin-import";
import tseslint from "typescript-eslint";
import testingLibrary from "eslint-plugin-testing-library";
import prettier from "eslint-config-prettier";

export default tseslint.config(
  {
    ignores: ["dist", "node_modules"],
  },

  // 1. Base Recommended Configs
  js.configs.recommended,
  ...tseslint.configs.recommended, // plugin:@typescript-eslint/recommended

  {
    files: ["**/*.{ts,tsx}"],
    plugins: {
      react,
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
      import: importPlugin,
      "testing-library": testingLibrary,
    },
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2020,
      },
      parserOptions: {
        warnOnUnsupportedTypeScriptVersion: false,
      },
    },
    settings: {
      react: { version: "detect" },
    },
    rules: {
      // --- React Hooks ---
      "no-undef": "off",
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "warn",
      "react/react-in-jsx-scope": "off",

      // --- TypeScript Loosening (as requested) ---
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-non-null-assertion": "off",
      "@typescript-eslint/no-unused-vars": "off",
      "@typescript-eslint/no-var-requires": "off",
      "@typescript-eslint/ban-ts-comment": "off",
      "@typescript-eslint/ban-types": "off",

      // --- Consistent Type Imports ---
      "@typescript-eslint/consistent-type-imports": [
        "error",
        {
          fixStyle: "inline-type-imports",
        },
      ],

      // --- Import Rules ---
      "import/no-duplicates": "error",
      "import/no-unresolved": "off", // Loosening this to prevent resolver crashes

      // --- Restricted Imports ---
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "lodash",
              message: "Import specific methods from `lodash`. e.g. `import map from 'lodash/map'`",
            },
            {
              name: "lodash-es",
              importNames: ["default"],
              message:
                "Import specific methods from `lodash-es`. e.g. `import { map } from 'lodash-es'`",
            },
            {
              name: "carbon-components-react",
              message:
                "Import from `@carbon/react` directly. e.g. `import { Toggle } from '@carbon/react'`",
            },
            {
              name: "@carbon/icons-react",
              message:
                "Import from `@carbon/react/icons`. e.g. `import { ChevronUp } from '@carbon/react/icons'`",
            },
          ],
        },
      ],

      // --- Testing Library (Base) ---
      ...testingLibrary.configs["flat/react"].rules,
    },
  },

  // 2. Overrides for E2E files
  {
    files: ["e2e/**/*.spec.ts"],
    rules: {
      "testing-library/prefer-screen-queries": "off",
    },
  },

  // 3. Prettier
  prettier,
);
