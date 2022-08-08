export interface EnumProps {
    label: string; // 选项框显示的文字
    value: any; // 选项框值
    disabled?: boolean; // 是否禁用此选项
    tagType?: string; // 当 tag 为 true 时，此选择会指定 tag 显示类型
    children?: EnumProps[]; // 为树形选择时，可以通过 children 属性指定子选项
    [key: string]: any;
}

export type SearchType =
    | 'text'
    | 'select'
    | 'multipleSelect'
    | 'treeSelect'
    | 'multipleTreeSelect'
    | 'date'
    | 'daterange'
    | 'timerange'
    | 'datetimerange';

export type TypeProp = 'index' | 'selection' | 'expand';

export type FixedProp = 'left' | 'right';

export interface ColumnProps {
    type: TypeProp; // index | selection | expand（特殊类型）
    prop: string; // 单元格数据（非特殊类型必填）
    label: string; // 单元格标题（非特殊类型必填）
    width: number | string; // 列宽
    minWidth: number | string; // 最小列宽
    isShow: boolean; // 是否显示在表格当中
    sortable: boolean; // 是否可排序（静态排序）
    fixed: FixedProp; // 固定列
    tag: boolean; // 是否是标签展示
    image: boolean; // 是否是图片展示
    search: boolean; // 是否为搜索项
    searchType: SearchType; // 搜索项类型
    searchProps: { [key: string]: any }; // 搜索项参数，根据 element 文档来，标签自带属性 > props 属性
    searchInitParam: string | number | boolean | any[]; // 搜索项初始值
    enum: EnumProps[] | (() => Promise<any>); // 枚举类型（渲染值的字典）
    renderHeader: (params: any) => any; // 自定义表头
}
