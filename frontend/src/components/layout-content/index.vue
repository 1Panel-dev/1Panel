<template>
    <div class="main-box">
        <div class="content-container__app" v-if="slots.app">
            <slot name="app"></slot>
        </div>
        <div class="content-container__toolbar" v-if="slots.toolbar">
            <slot name="toolbar"></slot>
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
                            <span class="float-right">
                                <slot v-if="slots.rightButton" name="rightButton"></slot>
                            </span>
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
    margin-top: 20px;
}

.content-container__search {
    margin-top: 20px;
    .el-card {
        --el-card-padding: 12px;
    }
}

.content-container__title {
    font-weight: 700;
    font-size: 18px;
}

.content-container__toolbar {
    margin-top: 20px;
    .el-button + .el-button {
        margin: 0 !important;
    }
    .el-button-group .el-button {
        margin: 0 !important;
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
    margin-top: 20px;
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
    margin-top: 20px;
}
</style>
