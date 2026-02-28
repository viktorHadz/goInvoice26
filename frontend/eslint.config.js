"use strict";
var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
    if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
        if (ar || !(i in from)) {
            if (!ar) ar = Array.prototype.slice.call(from, 0, i);
            ar[i] = from[i];
        }
    }
    return to.concat(ar || Array.prototype.slice.call(from));
};
Object.defineProperty(exports, "__esModule", { value: true });
var config_1 = require("eslint/config");
var eslint_config_typescript_1 = require("@vue/eslint-config-typescript");
var eslint_plugin_vue_1 = require("eslint-plugin-vue");
var eslint_plugin_oxlint_1 = require("eslint-plugin-oxlint");
var flat_1 = require("eslint-config-prettier/flat");
// To allow more languages other than `ts` in `.vue` files, uncomment the following lines:
// import { configureVueProject } from '@vue/eslint-config-typescript'
// configureVueProject({ scriptLangs: ['ts', 'tsx'] })
// More info at https://github.com/vuejs/eslint-config-typescript/#advanced-setup
exports.default = eslint_config_typescript_1.defineConfigWithVueTs.apply(void 0, __spreadArray(__spreadArray(__spreadArray(__spreadArray([{
        name: 'app/files-to-lint',
        files: ['**/*.{vue,ts,mts,tsx}'],
    },
    (0, config_1.globalIgnores)(['**/dist/**', '**/dist-ssr/**', '**/coverage/**'])], eslint_plugin_vue_1.default.configs['flat/essential'], false), [eslint_config_typescript_1.vueTsConfigs.recommended], false), eslint_plugin_oxlint_1.default.buildFromOxlintConfigFile('.oxlintrc.json'), false), [flat_1.default], false));
