import { PropType, ExtractPropTypes } from 'vue';
declare const _default: import("vue").DefineComponent<{
    title: {
        type: PropType<string>;
        default: string;
    };
    modalTitle: {
        type: PropType<string>;
        default: string;
    };
    visible: {
        type: PropType<boolean>;
    };
    width: {
        type: PropType<string>;
        default: string;
    };
    height: {
        type: PropType<string>;
        default: string;
    };
    trigger: {
        type: PropType<string | JSX.Element>;
    };
    onClick: {
        type: PropType<() => void>;
    };
    onClose: {
        type: PropType<() => void>;
    };
    /**
     * 显示全屏按钮
     */
    showAdjust: {
        type: PropType<boolean>;
        default: boolean;
    };
    isFullscreen: {
        type: PropType<boolean>;
        default: boolean;
    };
    onAdjust: {
        type: PropType<(val: boolean) => void>;
    };
}, () => JSX.Element, unknown, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, import("vue").EmitsOptions, "onClick" | "onClose" | "onAdjust", import("vue").VNodeProps & import("vue").AllowedComponentProps & import("vue").ComponentCustomProps, Readonly<ExtractPropTypes<{
    title: {
        type: PropType<string>;
        default: string;
    };
    modalTitle: {
        type: PropType<string>;
        default: string;
    };
    visible: {
        type: PropType<boolean>;
    };
    width: {
        type: PropType<string>;
        default: string;
    };
    height: {
        type: PropType<string>;
        default: string;
    };
    trigger: {
        type: PropType<string | JSX.Element>;
    };
    onClick: {
        type: PropType<() => void>;
    };
    onClose: {
        type: PropType<() => void>;
    };
    /**
     * 显示全屏按钮
     */
    showAdjust: {
        type: PropType<boolean>;
        default: boolean;
    };
    isFullscreen: {
        type: PropType<boolean>;
        default: boolean;
    };
    onAdjust: {
        type: PropType<(val: boolean) => void>;
    };
}>>, {
    title: string;
    modalTitle: string;
    width: string;
    height: string;
    showAdjust: boolean;
    isFullscreen: boolean;
}>;
export default _default;
