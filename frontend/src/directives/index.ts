import { App, Directive } from 'vue';
import integerInput from './modules/integer';
import permission from './modules/permission';

const directivesList: { [key: string]: Directive } = {
    'integer-input': integerInput,
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
