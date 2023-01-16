<template>
    <div class="main-box">
        <div class="content-container__header" v-if="slots.header">
            <slot name="header"></slot>
        </div>
        <div class="content-container__app" v-if="slots.app">
            <slot name="app"></slot>
        </div>
        <div class="content-container__toolbar" v-if="slots.toolbar">
            <slot name="toolbar"></slot>
        </div>
        <div class="content-container_form">
            <slot name="form">
                <form-button>
                    <slot name="button"></slot>
                </form-button>
            </slot>
        </div>
        <div class="content-container__main" v-if="slots.main">
            <el-card>
                <div class="content-container__title" v-if="slots.title || title">
                    <slot name="title">
                        <back-button
                            :path="backPath"
                            :name="backName"
                            :to="backTo"
                            :header="title"
                            :reload="reload"
                            v-if="showBack"
                        >
                            <template v-if="slots.buttons" #buttons>
                                <slot name="buttons"></slot>
                            </template>
                        </back-button>

                        <span v-else>
                            {{ title }}
                            <span v-if="slots.buttons">
                                <el-divider direction="vertical" />
                                <slot name="buttons"></slot>
                            </span>
                            <span style="float: right">
                                <slot v-if="slots.rightButton" name="rightButton"></slot>
                            </span>
                        </span>
                    </slot>
                </div>
                <div v-if="slots.prompt">
                    <slot name="prompt"></slot>
                </div>
                <slot name="main"></slot>
            </el-card>
        </div>
        <slot></slot>
    </div>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue';
import BackButton from '@/components/back-button/index.vue';
import FormButton from './form-button.vue';
defineOptions({ name: 'LayoutContent' });
const slots = useSlots();
const prop = defineProps({
    title: String,
    backPath: String,
    backName: String,
    backTo: Object,
    reload: Boolean,
});

const showBack = computed(() => {
    const { backPath, backName, backTo, reload } = prop;
    return backPath || backName || backTo || reload;
});
</script>

<style lang="scss">
@use '@/styles/mixins.scss' as *;

.content-container__search {
    margin-top: 20px;
}

.content-container__title {
    font-weight: 700;
    font-size: 18px;
}

.content-container__toolbar {
    // @include flex-row(space-between, center);
    margin-top: 30px;
}

.content-container_form {
    text-align: -webkit-center;
    width: 60%;
    margin-left: 15%;
    .form-button {
        float: right;
    }
}

.content-container__main {
    margin-top: 20px;
}

.el-divider--horizontal {
    display: block;
    height: 1px;
    width: 100%;
    margin: 10px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
</style>
