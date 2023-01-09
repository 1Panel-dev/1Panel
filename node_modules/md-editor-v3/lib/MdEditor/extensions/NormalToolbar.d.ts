import { PropType, ExtractPropTypes } from 'vue';
declare const _default: import("vue").DefineComponent<{
    title: {
        type: PropType<string>;
        default: string;
    };
    trigger: {
        type: PropType<string | JSX.Element>;
    };
    onClick: {
        type: PropType<(e: MouseEvent) => void>;
    };
}, () => JSX.Element, unknown, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, import("vue").EmitsOptions, "onClick", import("vue").VNodeProps & import("vue").AllowedComponentProps & import("vue").ComponentCustomProps, Readonly<ExtractPropTypes<{
    title: {
        type: PropType<string>;
        default: string;
    };
    trigger: {
        type: PropType<string | JSX.Element>;
    };
    onClick: {
        type: PropType<(e: MouseEvent) => void>;
    };
}>>, {
    title: string;
}>;
export default _default;
