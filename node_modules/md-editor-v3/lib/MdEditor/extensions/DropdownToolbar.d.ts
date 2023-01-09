import { PropType, ExtractPropTypes } from 'vue';
declare const _default: import("vue").DefineComponent<{
    title: {
        type: PropType<string>;
        default: string;
    };
    visible: {
        type: PropType<boolean>;
    };
    trigger: {
        type: PropType<string | JSX.Element>;
    };
    onChange: {
        type: PropType<(visible: boolean) => void>;
    };
    overlay: {
        type: PropType<string | JSX.Element>;
    };
}, () => JSX.Element, unknown, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, import("vue").EmitsOptions, "onChange", import("vue").VNodeProps & import("vue").AllowedComponentProps & import("vue").ComponentCustomProps, Readonly<ExtractPropTypes<{
    title: {
        type: PropType<string>;
        default: string;
    };
    visible: {
        type: PropType<boolean>;
    };
    trigger: {
        type: PropType<string | JSX.Element>;
    };
    onChange: {
        type: PropType<(visible: boolean) => void>;
    };
    overlay: {
        type: PropType<string | JSX.Element>;
    };
}>>, {
    title: string;
}>;
export default _default;
