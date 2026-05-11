<template>
    <div ref="containerRef" class="tag-container flex gap-2">
        <el-check-tag :checked="activeTag == 'all'" @click="changeTag('all')">
            {{ $t('app.all') }}
        </el-check-tag>
        <el-check-tag
            :checked="activeTag == item.key"
            v-for="item in visibleTags"
            :key="item.key"
            @click="changeTag(item.key)"
        >
            {{ item.name }}
        </el-check-tag>
        <div class="more-tag inline" v-if="hiddenTags.length > 0">
            <el-dropdown>
                <el-check-tag :checked="moreTag !== ''">
                    {{ moreLabel }}
                    <el-icon :size="10">
                        <arrow-down />
                    </el-icon>
                </el-check-tag>
                <template #dropdown>
                    <el-dropdown-menu>
                        <el-dropdown-item v-for="item in hiddenTags" @click="changeTag(item.key)" :key="item.key">
                            {{ item.name }}
                        </el-dropdown-item>
                    </el-dropdown-menu>
                </template>
            </el-dropdown>
        </div>
    </div>
    <div class="tag-measurements" aria-hidden="true">
        <el-check-tag ref="allMeasureRef">{{ $t('app.all') }}</el-check-tag>
        <el-check-tag v-for="item in tags" :key="`measure-${item.key}`" :ref="(el) => setMeasureTagRef(item.key, el)">
            {{ item.name }}
        </el-check-tag>
        <el-check-tag ref="moreMeasureRef">
            {{ moreLabel }}
            <el-icon :size="10">
                <arrow-down />
            </el-icon>
        </el-check-tag>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { getAppTags } from '@/api/modules/app';
import { App } from '@/api/interface/app';

const props = defineProps({
    hideKey: {
        type: String,
        default: '',
    },
});

const containerRef = ref<HTMLElement>();
const activeTag = ref('all');
const tags = ref<App.Tag[]>([]);
const moreTag = ref('');
const visibleTagCount = ref(7);
const allMeasureRef = ref();
const moreMeasureRef = ref();
const emit = defineEmits(['change']);
const measureTagRefs: Record<string, any> = {};

const visibleTags = computed(() => tags.value.slice(0, visibleTagCount.value));
const hiddenTags = computed(() => tags.value.slice(visibleTagCount.value));
const moreLabel = computed(() => {
    if (moreTag.value === '') {
        return i18n.global.t('tabs.more');
    }
    return getTagValue(moreTag.value) || i18n.global.t('tabs.more');
});

const getTagValue = (key: string) => {
    const tag = tags.value.find((tag) => tag.key === key);
    if (tag) {
        return tag.name;
    }
    return '';
};

const setMeasureTagRef = (key: string, el: any) => {
    if (el) {
        measureTagRefs[key] = el;
        return;
    }
    delete measureTagRefs[key];
};

const getElementWidth = (target: any) => {
    const el = target?.$el || target;
    if (!(el instanceof HTMLElement)) {
        return 0;
    }
    return el.getBoundingClientRect().width;
};

const getContainerGap = () => {
    if (!containerRef.value) {
        return 0;
    }
    const styles = window.getComputedStyle(containerRef.value);
    const gap = Number.parseFloat(styles.columnGap || styles.gap || '0');
    return Number.isNaN(gap) ? 0 : gap;
};

const changeTag = (key: string) => {
    activeTag.value = key;
    emit('change', key);
    const index = tags.value.findIndex((tag) => tag.key === key);
    if (index >= visibleTagCount.value) {
        moreTag.value = key;
    } else {
        moreTag.value = '';
    }
    nextTick(() => {
        calculateVisibleTagCount();
    });
};

const calculateVisibleTagCount = () => {
    if (!containerRef.value) return;

    const safeReserve = 8;
    const containerWidth = containerRef.value.getBoundingClientRect().width - safeReserve;
    const gap = getContainerGap();
    const allWidth = getElementWidth(allMeasureRef.value);
    const moreWidth = getElementWidth(moreMeasureRef.value);

    if (!containerWidth || !allWidth) {
        return;
    }

    let usedWidth = allWidth;
    let count = 0;

    for (let i = 0; i < tags.value.length; i++) {
        const currentWidth = getElementWidth(measureTagRefs[tags.value[i].key]);
        if (!currentWidth) {
            continue;
        }

        const hasRemainingTags = i < tags.value.length - 1;
        const nextWidth = usedWidth + gap + currentWidth + (hasRemainingTags ? gap + moreWidth : 0);

        if (nextWidth <= containerWidth) {
            usedWidth += gap + currentWidth;
            count++;
            continue;
        }

        break;
    }

    visibleTagCount.value = count;

    const activeIndex = tags.value.findIndex((tag) => tag.key === activeTag.value);
    const expectedMoreTag = activeIndex >= visibleTagCount.value ? activeTag.value : '';
    if (expectedMoreTag !== moreTag.value) {
        moreTag.value = expectedMoreTag;
        nextTick(() => {
            calculateVisibleTagCount();
        });
    }
};

let resizeObserver: ResizeObserver | null = null;

const initResizeObserver = () => {
    if (!containerRef.value) return;

    resizeObserver = new ResizeObserver(() => {
        calculateVisibleTagCount();
    });

    resizeObserver.observe(containerRef.value);
};

const getTags = async () => {
    await getAppTags().then((res) => {
        for (let i = 0; i < res.data.length; i++) {
            if (res.data[i].key === props.hideKey) {
                res.data.splice(i, 1);
                break;
            }
        }
        tags.value = res.data;
        nextTick(() => {
            calculateVisibleTagCount();
        });
    });
};

onMounted(() => {
    getTags();
    nextTick(() => {
        initResizeObserver();
    });
});

onBeforeUnmount(() => {
    if (resizeObserver) {
        resizeObserver.disconnect();
    }
});
</script>

<style lang="scss" scoped>
.tag-container {
    flex-wrap: nowrap;
    align-items: center;
    overflow: hidden;
}

.more-tag {
    flex-shrink: 0;
}

.tag-measurements {
    position: absolute;
    visibility: hidden;
    pointer-events: none;
    overflow: hidden;
    white-space: nowrap;
    height: 0;
}

.tag-container :deep(.el-check-tag),
.tag-measurements :deep(.el-check-tag) {
    flex-shrink: 0;
    white-space: nowrap;
}

.el-check-tag.el-check-tag--primary.is-checked {
    background-color: var(--el-color-info-light-9) !important;
}

.el-check-tag:hover {
    background-color: var(--el-color-info-light-9) !important;
}
</style>
