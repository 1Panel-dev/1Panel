<template>
    <div class="content-box">
        <el-alert
            title="通过 component :is 组件属性 && v-bind 属性透传，可以将 template 中的 html 代码全部改变为 columns 配置项，具体配置请看代码。"
            type="warning"
            :closable="false"
        />
        <component
            :is="'el-form'"
            v-bind="options.form"
            ref="proFormRef"
            :model="model"
        >
            <template v-for="item in options.columns" :key="item.prop">
                <component :is="'el-form-item'" v-bind="item.formItem">
                    <component
                        :is="`el-${item.attrs.typeName}`"
                        v-bind="item.attrs"
                        v-model="model[item.formItem.prop]"
                    />
                </component>
            </template>
            <el-form-item>
                <slot name="operation"></slot>
            </el-form-item>
        </component>
    </div>
</template>

<script setup lang="ts" name="proForm">
import { ref } from 'vue';

let model = ref<any>({});

const options = ref({
    // 表单整体配置项
    form: {
        inline: false,
        labelPosition: 'right',
        labelWidth: '80px',
        size: 'default',
        disabled: false,
        labelSuffix: ' :',
    },
    // 表单列配置项 (formItem 代表 item 配置项，attrs 代表 输入、选择框 配置项)
    columns: [
        {
            formItem: {
                label: '用户名',
                prop: 'username',
                labelWidth: '80px',
                required: true,
            },
            attrs: {
                typeName: 'input',
                clearable: true,
                placeholder: '请输入用户名',
                disabled: true,
            },
        },
        {
            formItem: {
                label: '密码',
                prop: 'password',
                class: 'data',
            },
            attrs: {
                typeName: 'input',
                clearable: true,
                autofocus: true,
                placeholder: '请输入密码',
                type: 'password',
            },
        },
        {
            formItem: {
                label: '邮箱',
                prop: 'email',
            },
            attrs: {
                typeName: 'input',
                placeholder: '请输入邮箱',
                clearable: true,
                style: 'width:500px',
            },
        },
    ],
});
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
