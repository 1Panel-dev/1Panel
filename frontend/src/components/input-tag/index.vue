<template>
    <div>
        <el-input-tag
            v-model="tmpTags"
            trigger="Enter"
            @paste="handlePaste"
            @change="handleUpdate"
            @add-tag="handleAdd"
        >
            <template #suffix>
                <el-tooltip :content="$t('commons.button.copy')">
                    <el-button class="-mr-3" link icon="CopyDocument" @click="copyText(tmpTags.join('\n'))" />
                </el-tooltip>
                <el-tooltip :content="$t('commons.button.clean')">
                    <el-button link icon="Close" @click="handleClean" />
                </el-tooltip>
            </template>
        </el-input-tag>
        <span class="input-help">{{ $t('commons.rule.inputTags', [props.egHelp]) }}</span>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { copyText } from '@/utils/util';

const em = defineEmits(['update:tags']);
const tmpTags = ref([]);
const props = defineProps({
    egHelp: { type: String, default: 'key=val' },
    tags: { type: Array<string>, default: [] },
});

watch(
    () => props.tags,
    (newVal) => {
        tmpTags.value = newVal || [];
    },
);

const handlePaste = (event: any) => {
    event.preventDefault();
    const pasteData = event.clipboardData.getData('text');
    const tags = pasteData.split('\n');
    for (const item of tags) {
        if (item) {
            handleAdd(item);
        }
    }
};
const handleAdd = (val: string) => {
    tmpTags.value = tmpTags.value?.filter((item) => item !== val);
    tmpTags.value.push(val);
};
const handleUpdate = () => {
    em('update:tags', tmpTags.value);
};
const handleClean = () => {
    tmpTags.value = [];
    handleUpdate();
};

onMounted(() => {
    tmpTags.value = props.tags || [];
});
</script>
