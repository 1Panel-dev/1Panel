<template>
    <div>
        <el-button
            class="tag-button"
            :class="activeTag === 'all' ? '' : 'no-active'"
            @click="changeTag('all')"
            :type="activeTag === 'all' ? 'primary' : ''"
            :plain="activeTag !== 'all'"
        >
            {{ $t('app.all') }}
        </el-button>
        <div v-for="item in tags.slice(0, 7)" :key="item.key" class="inline">
            <el-button
                class="tag-button"
                :class="activeTag === item.key ? '' : 'no-active'"
                @click="changeTag(item.key)"
                :type="activeTag === item.key ? 'primary' : ''"
                :plain="activeTag !== item.key"
            >
                {{ item.name }}
            </el-button>
        </div>
        <div class="inline">
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
                        <el-dropdown-item v-for="item in tags.slice(7)" @click="changeTag(item.key)" :key="item.key">
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

const activeTag = ref('all');
const tags = ref<App.Tag[]>([]);
const moreTag = ref('');
const emit = defineEmits(['change']);

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
    if (index > 6) {
        moreTag.value = key;
    } else {
        moreTag.value = '';
    }
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
    });
};

onMounted(() => {
    getTags();
});
</script>
