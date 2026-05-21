<template>
    <el-drawer
        v-model="localOpenPage"
        @close="handleClose"
        :destroy-on-close="true"
        :before-close="beforeClose"
        append-to-body
        :size="isFull ? '100%' : size"
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
                        <el-button
                            @click="toggleFullscreen"
                            link
                            icon="FullScreen"
                            plain
                            class="-mt-1 mr-2"
                        ></el-button>
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
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isFullScreen } = useGlobalStore();

const isFull = ref();

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
    confirmBeforeClose: {
        type: Boolean,
        default: false,
    },
});

const slots = useSlots();
const emit = defineEmits(['update:modelValue', 'close', 'beforeClose']);

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
    if (props.confirmBeforeClose) {
        const done = () => {
            if (props.back) {
                props.back();
            } else {
                localOpenPage.value = false;
                isFullScreen.value = false;
            }
        };
        emit('beforeClose', done);
    } else {
        if (props.back) {
            props.back();
        } else {
            localOpenPage.value = false;
            isFullScreen.value = false;
        }
    }
};

const handleClose = () => {
    localOpenPage.value = false;
    isFullScreen.value = false;
    emit('close');
};

const beforeClose = (done: () => void) => {
    if (!props.confirmBeforeClose) {
        done();
    } else {
        emit('beforeClose', done);
    }
};

function toggleFullscreen() {
    isFullScreen.value = !isFullScreen.value;
    isFull.value = isFullScreen.value;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullScreen.value ? 'quitFullscreen' : 'fullscreen'));
};
</script>
