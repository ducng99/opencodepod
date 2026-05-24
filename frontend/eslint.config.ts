import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import pluginReact from "eslint-plugin-react";
import { defineConfig, globalIgnores } from "eslint/config";
import stylistic from "@stylistic/eslint-plugin";

export default defineConfig([
    globalIgnores(["dist/**"]),
    { files: ["**/*.{js,mjs,cjs,ts,mts,cts,jsx,tsx}"], plugins: { js }, extends: ["js/recommended"], languageOptions: { globals: globals.browser } },
    tseslint.configs.recommended,
    pluginReact.configs.flat.recommended,
    stylistic.configs.customize({
        indent: 4,
        quotes: "double",
        semi: true,
        jsx: true,
    }),
    {
        rules: {
            "react/react-in-jsx-scope": "off",
        },
    },
]);
