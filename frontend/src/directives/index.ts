import { App, Directive } from 'vue';
import integerInput from './modules/integer';
import permission, { nodeAdminDirective } from './modules/permission';

const directivesList: { [key: string]: Directive } = {
    'integer-input': integerInput,
    'node-admin': nodeAdminDirective,
    permission,
};

const directives = {
    install: function (app: App<Element>) {
        Object.keys(directivesList).forEach((key) => {
            app.directive(key, directivesList[key]);
        });
    },
};

export default directives;
