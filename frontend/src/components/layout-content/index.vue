<template>
    <div class="main-box">
        <div class="content-container__app" v-if="slots.app">
            <slot name="app"></slot>
        </div>

        <div class="content-container__search" v-if="slots.search">
            <el-card>
                <slot name="search"></slot>
            </el-card>
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
                <div class="content-container__title">
                    <slot name="title">
                        <div v-if="showBack" class="flex flex-wrap gap-4 sm:justify-between">
                            <back-button
                                :path="backPath"
                                :name="backName"
                                :to="backTo"
                                :header="title"
                                :reload="reload"
                            >
                                <template v-if="slots.leftToolBar" #buttons>
                                    <div class="flex flex-wrap gap-2 items-center justify-start">
                                        <slot name="leftToolBar" v-if="slots.leftToolBar"></slot>
                                    </div>
                                </template>
                            </back-button>
                            <div class="flex flex-wrap gap-3">
                                <slot name="rightToolBar" v-if="slots.rightToolBar"></slot>
                            </div>
                        </div>
                        <div v-else class="flex flex-wrap gap-4 sm:justify-between">
                            <div class="flex gap-2 flex-wrap items-center justify-start">
                                <slot name="leftToolBar" v-if="slots.leftToolBar"></slot>
                            </div>
                            <div class="flex flex-wrap gap-3" v-if="slots.rightToolBar">
                                <slot name="rightToolBar"></slot>
                            </div>
                        </div>

                        <span v-if="slots.toolbar">
                            <slot name="toolbar"></slot>
                        </span>

                        <div v-if="prop.divider">
                            <div class="divider"></div>
                        </div>
                    </slot>
                </div>
                <div v-if="slots.prompt" class="prompt">
                    <slot name="prompt"></slot>
                </div>
                <div class="main-content">
                    <slot name="main"></slot>
                </div>
            </el-card>
        </div>
        <slot></slot>
    </div>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue';
import FormButton from './form-button.vue';
defineOptions({ name: 'LayoutContent' });
const slots = useSlots();
const prop = defineProps({
    title: String,
    backPath: String,
    backName: String,
    backTo: Object,
    reload: Boolean,
    divider: Boolean,
});

const showBack = computed(() => {
    const { backPath, backName, backTo, reload } = prop;
    return backPath || backName || backTo || reload;
});
</script>

<style lang="scss">
@use '@/styles/mixins.scss' as *;

.content-container__app {
    margin-top: 7px;
}

.content-container__search {
    margin-top: 7px;
    .el-card {
        --el-card-padding: 12px;
    }
}

.content-container__title {
    font-weight: 400;
    font-size: 18px;
    .el-button + .el-button {
        margin: 0 !important;
    }
    .el-button-group > .el-button + .el-button {
        margin-left: 0 !important;
    }
    .el-button-group > .el-button:not(:last-child) {
        margin-right: -1px !important;
    }
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
    margin-top: 10px;
}

.prompt {
    margin-top: 10px;
}

.divider {
    margin-top: 20px;
    border: 0;
    border-top: var(--panel-border);
}

.main-box {
    position: relative;
}
.main-content {
    margin-top: 15px;
}
</style>
