<template>
    <div ref="containerRef" class="flex gap-2">
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
        <div class="inline" v-if="hiddenTags.length > 0">
            <el-dropdown>
                <el-button
                    class="tag-button"
                    :type="moreTag !== '' ? 'primary' : ''"
                    :class="moreTag !== '' ? '' : 'no-active'"
                >
                    {{ moreTag == '' ? $t('tabs.more') : getTagValue(moreTag) }}
                    <el-icon class="el-icon--right">
                        <arrow-down />
                    </el-icon>
                </el-button>
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
</template>

<script lang="ts" setup>
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
const emit = defineEmits(['change']);

const visibleTags = computed(() => tags.value.slice(0, visibleTagCount.value));
const hiddenTags = computed(() => tags.value.slice(visibleTagCount.value));

const getTagValue = (key: string) => {
    const tag = tags.value.find((tag) => tag.key === key);
    if (tag) {
        return tag.name;
    }
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
};

const calculateVisibleTagCount = () => {
    if (!containerRef.value) return;

    const containerWidth = containerRef.value.offsetWidth;

    if (containerWidth >= 1800) {
        visibleTagCount.value = 18;
    } else if (containerWidth >= 1400) {
        visibleTagCount.value = 15;
    } else if (containerWidth >= 1200) {
        visibleTagCount.value = 12;
    } else if (containerWidth >= 992) {
        visibleTagCount.value = 8;
    } else if (containerWidth >= 768) {
        visibleTagCount.value = 6;
    } else if (containerWidth >= 576) {
        visibleTagCount.value = 4;
    } else {
        visibleTagCount.value = 2;
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
.el-check-tag.el-check-tag--primary.is-checked {
    background-color: var(--el-color-info-light-9) !important;
}
</style>
