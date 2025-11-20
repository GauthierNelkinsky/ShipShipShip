import js from "@eslint/js";
import ts from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import svelte from "eslint-plugin-svelte";
import svelteParser from "svelte-eslint-parser";
import globals from "globals";

export default [
  js.configs.recommended,
  {
    files: ["**/*.js", "**/*.ts"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 2022,
        sourceType: "module",
        extraFileExtensions: [".svelte"],
      },
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2021,
        NodeJS: "readonly",
        RequestInit: "readonly",
      },
    },
    plugins: {
      "@typescript-eslint": ts,
    },
    rules: {
      ...ts.configs.recommended.rules,
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_|^\\$\\$",
        },
      ],
      "@typescript-eslint/no-explicit-any": "warn",
      "no-console": ["warn", { allow: ["warn", "error"] }],
      "no-undef": "off", // TypeScript handles this
    },
  },
  {
    files: ["**/*.svelte"],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: tsParser,
        ecmaVersion: 2022,
        sourceType: "module",
        extraFileExtensions: [".svelte"],
      },
      globals: {
        ...globals.browser,
        $$Props: "readonly",
        $$Events: "readonly",
        $$Slots: "readonly",
      },
    },
    plugins: {
      svelte,
      "@typescript-eslint": ts,
    },
    rules: {
      ...svelte.configs.recommended.rules,
      "svelte/no-unused-svelte-ignore": "warn",
      "svelte/valid-compile": "warn",
      // Handle unused vars in Svelte files
      "no-unused-vars": "off",
      "no-undef": "off",
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_|^\\$\\$Props|^\\$\\$Events|^\\$\\$Slots",
        },
      ],
      "no-self-assign": "warn",
    },
  },
  {
    // Config files
    files: ["*.config.js", "*.config.ts"],
    rules: {
      "@typescript-eslint/no-unused-vars": "off",
      "@typescript-eslint/no-require-imports": "off",
    },
  },
  {
    ignores: [
      ".svelte-kit/**",
      "build/**",
      "node_modules/**",
      ".DS_Store",
      "dist/**",
      "coverage/**",
      "*.log",
    ],
  },
];
