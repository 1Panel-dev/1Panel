<template>
    <el-drawer
        v-model="localOpenPage"
        @close="handleClose"
        :destroy-on-close="true"
        :size="size"
        :close-on-press-escape="autoClose"
        :close-on-click-modal="autoClose"
    >
        <template #header>
            <el-page-header @back="handleBack">
                <template #content>
                    <span>{{ header }}</span>
                    <span v-if="resource != ''">
                        -
                        <el-tooltip v-if="resource.length > 25" :content="resource" placement="bottom">
                            <el-text type="primary">{{ resource.substring(0, 23) + '...' }}</el-text>
                        </el-tooltip>
                        <el-text type="primary" v-else>{{ resource }}</el-text>
                    </span>
                    <el-divider v-if="slots.buttons" direction="vertical" />
                    <slot v-if="slots.buttons" name="buttons"></slot>
                </template>
                <template #extra>
                    <el-tooltip :content="loadTooltip()" placement="top" v-if="fullScreen">
                        <el-button @click="toggleFullscreen" link icon="FullScreen" plain class="mr-5"></el-button>
                    </el-tooltip>
                    <slot v-if="slots.extra" name="extra"></slot>
                </template>
            </el-page-header>
        </template>

        <div ref="drawerContent">
            <div v-if="slots.content">
                <slot name="content"></slot>
            </div>
            <el-row v-else>
                <el-col :span="22" :offset="1">
                    <slot></slot>
                </el-col>
            </el-row>
        </div>

        <template #footer v-if="slots.footer">
            <slot name="footer"></slot>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { computed, useSlots, ref } from 'vue';
defineOptions({ name: 'DrawerPro' });
import i18n from '@/lang';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const drawerContent = ref();

const props = defineProps({
    header: String,
    back: Function,
    resource: {
        type: String,
        default: '',
    },
    size: {
        type: String,
        default: 'normal',
    },
    modelValue: {
        type: Boolean,
        default: false,
    },
    fullScreen: {
        type: Boolean,
        default: false,
    },
    autoClose: {
        type: Boolean,
        default: true,
    },
});

const slots = useSlots();
const emit = defineEmits(['update:modelValue', 'close']);

const size = computed(() => {
    switch (props.size) {
        case 'small':
            return '30%';
        case 'normal':
            return '40%';
        case 'large':
            return '50%';
        case 'full':
            return '100%';
        case '60%':
            return '60%';
        case props.size:
            return props.size;
        default:
            return '50%';
    }
});

const localOpenPage = computed({
    get() {
        return props.modelValue;
    },
    set(value: boolean) {
        emit('update:modelValue', value);
    },
});

const handleBack = () => {
    if (props.back) {
        props.back();
    } else {
        handleClose();
    }
};

const handleClose = () => {
    localOpenPage.value = false;
    globalStore.isFullScreen = false;
    emit('close');
};

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};
</script>
