import { marked, Renderer, Slugger } from 'marked';
declare global {
    interface Window {
        hljs: any;
        prettier: any;
        prettierPlugins: any;
        Cropper: any;
        screenfull: any;
        mermaid: any;
        katex: any;
    }
}
export interface ToolbarTips {
    bold?: string;
    underline?: string;
    italic?: string;
    strikeThrough?: string;
    title?: string;
    sub?: string;
    sup?: string;
    quote?: string;
    unorderedList?: string;
    orderedList?: string;
    task?: string;
    codeRow?: string;
    code?: string;
    link?: string;
    image?: string;
    table?: string;
    mermaid?: string;
    katex?: string;
    revoke?: string;
    next?: string;
    save?: string;
    prettier?: string;
    pageFullscreen?: string;
    fullscreen?: string;
    preview?: string;
    htmlPreview?: string;
    catalog?: string;
    github?: string;
    '-'?: string;
    '='?: string;
}
export interface StaticTextDefaultValue {
    toolbarTips?: ToolbarTips;
    titleItem?: {
        h1?: string;
        h2?: string;
        h3?: string;
        h4?: string;
        h5?: string;
        h6?: string;
    };
    imgTitleItem?: {
        link: string;
        upload: string;
        clip2upload: string;
    };
    linkModalTips?: {
        linkTitle?: string;
        imageTitle?: string;
        descLabel?: string;
        descLabelPlaceHolder?: string;
        urlLabel?: string;
        urlLabelPlaceHolder?: string;
        buttonOK?: string;
    };
    clipModalTips?: {
        title?: string;
        buttonUpload?: string;
    };
    copyCode?: {
        text?: string;
        successTips?: string;
        failTips?: string;
    };
    mermaid?: {
        flow?: string;
        sequence?: string;
        gantt?: string;
        class?: string;
        state?: string;
        pie?: string;
        relationship?: string;
        journey?: string;
    };
    katex?: {
        inline: string;
        block: string;
    };
    footer?: {
        markdownTotal: string;
        scrollAuto: string;
    };
}
export interface StaticTextDefault {
    'zh-CN': StaticTextDefaultValue;
    'en-US': StaticTextDefaultValue;
}
export declare type StaticTextDefaultKey = keyof StaticTextDefault;
export declare type ToolbarNames = keyof ToolbarTips | number;
export declare type Footers = '=' | 'markdownTotal' | 'scrollSwitch' | number;
export interface SettingType {
    pageFullscreen: boolean;
    fullscreen: boolean;
    preview: boolean;
    htmlPreview: boolean;
}
export declare type Themes = 'light' | 'dark';
/**
 * 预览主题
 *
 * @list ['default', 'github', 'vuepress', 'mk-cute', 'smart-blue', 'cyanosis']
 */
export declare type PreviewThemes = string;
export interface HeadList {
    text: string;
    level: 1 | 2 | 3 | 4 | 5 | 6;
    active?: boolean;
}
export declare type MarkedHeadingId = (text: string, level: number, index: number) => string;
export interface MermaidTemplate {
    /**
     * 流程图
     */
    flow?: string;
    /**
     * 时序图
     */
    sequence?: string;
    /**
     * 甘特图
     */
    gantt?: string;
    /**
     * 类图
     */
    class?: string;
    /**
     * 状态图
     */
    state?: string;
    /**
     * 饼图
     */
    pie?: string;
    /**
     * 关系图
     */
    relationship?: string;
    /**
     * 旅程图
     */
    journey?: string;
}
export declare type RewriteHeading = (text: string, level: 1 | 2 | 3 | 4 | 5 | 6, raw: string, slugger: Slugger, index: number) => string;
export interface RewriteRenderer extends Omit<Renderer, 'heading'> {
    heading: RewriteHeading;
}
export interface ConfigOption {
    /**
     * 覆盖编辑器默认的renderer属性
     * @see https://marked.js.org/using_pro#renderer
     */
    markedRenderer?: (renderer: RewriteRenderer) => RewriteRenderer;
    /**
     * 自定义 marked 扩展
     * @see https://marked.js.org/using_pro#extensions
     */
    markedExtensions?: Array<marked.TokenizerExtension & marked.RendererExtension>;
    /**
     * 自定义 marked option，不推荐在这么覆盖renderer，这会导致内部逻辑混乱！
     * @see https://marked.js.org/using_advanced#options
     */
    markedOptions?: marked.MarkedOptions;
    /**
     * 编辑器内部依赖库
     */
    editorExtensions?: {
        highlight?: {
            instance?: any;
            js?: string;
            css?: CodeCss;
        };
        prettier?: {
            prettierInstance?: any;
            parserMarkdownInstance?: any;
            standaloneJs?: string;
            parserMarkdownJs?: string;
        };
        cropper?: {
            instance?: any;
            js?: string;
            css?: string;
        };
        iconfont?: string;
        screenfull?: {
            instance?: any;
            js?: string;
        };
        mermaid?: {
            instance?: any;
            js?: string;
        };
        katex?: {
            instance?: any;
            js?: string;
            css?: string;
        };
    };
    editorConfig?: {
        /**
         * 自定义提示语言
         */
        languageUserDefined?: {
            [key: string]: StaticTextDefaultValue;
        };
        /**
         * 自定义内部mermaid模块
         */
        mermaidTemplate?: MermaidTemplate;
        /**
         * 输入渲染延迟（ms）
         */
        renderDelay?: number;
    };
}
/**
 * 扩展编辑器内部功能，包括marked和一些内部依赖实例，如highlight、cropper等
 */
export declare type Config = (options: ConfigOption) => void;
/**
 * 编辑器操作潜在的错误
 */
export interface InnerError {
    name: string;
    message: string;
}
export interface CodeCss {
    [key: string]: {
        light: string;
        dark: string;
    };
}
export declare type UpdateSetting = (k: keyof SettingType, v?: boolean) => void;
export declare type ChangeEvent = (v: string) => void;
export declare type SaveEvent = (v: string, h: Promise<string>) => void;
export declare type UploadImgEvent = (files: Array<File>, callBack: (urls: string[]) => void) => void;
export declare type HtmlChangedEvent = (h: string) => void;
export declare type GetCatalogEvent = (list: HeadList[]) => void;
export declare type ErrorEvent = (err: InnerError) => void;
export interface ExposeEvent {
    pageFullscreen(status: boolean): void;
    fullscreen(status: boolean): void;
    preview(status: boolean): void;
    htmlPreview(status: boolean): void;
    catalog(status: boolean): void;
}
export interface InsertParam {
    targetValue: string;
    select: boolean;
    deviationStart: number;
    deviationEnd: number;
}
export declare type InsertContentGenerator = (selectedText: string) => InsertParam;
export interface ExposeParam {
    /**
     * 添加事件监听
     *
     * @param eventName 事件名称
     * @param callBack 事件回调函数
     */
    on<E extends keyof ExposeEvent, C extends ExposeEvent[E]>(eventName: E, callBack: C): void;
    /**
     * 切换页面内全屏
     *
     * @param status 是否页面全屏
     */
    togglePageFullscreen(status?: boolean): void;
    /**
     * 切换屏幕全屏
     *
     * @param status 是否屏幕全屏
     */
    toggleFullscreen(status?: boolean): void;
    /**
     * 切换是否显示预览
     *
     * @param status 是否显示预览
     */
    togglePreview(status?: boolean): void;
    /**
     * 切换是否显示html预览
     *
     * @param status html预览状态
     */
    toggleHtmlPreview(status?: boolean): void;
    /**
     * 切换是否显示目录
     *
     * @param status 是否显示目录，不设置默认相反
     */
    toggleCatalog(status?: boolean): void;
    /**
     * 触发保存
     */
    triggerSave(): void;
    /**
     * 手动向文本框插入内容
     *
     * @param {Function} generate 构造插入内容方法
     * 构造方法提供「当前选中」的内容为入参
     * 返回「待插入内容」和插入的属性
     * 入参 selectedText 当前选中的内容
     *
     * targetValue     待插入内容
     * select         插入后是否自动选中内容
     * deviationStart 插入后选中位置的开始偏移量
     * deviationEnd   插入后选中位置的结束偏移量
     *
     */
    insert(generate: InsertContentGenerator): void;
    /**
     * 手动聚焦
     */
    focus(): void;
}
