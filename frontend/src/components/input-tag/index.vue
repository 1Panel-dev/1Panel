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
                <el-tooltip v-if="withFile" :content="$t('website.select')">
                    <el-button
                        class="-mr-3"
                        link
                        icon="Folder"
                        @click="fileRef.acceptParams({ path: baseDir, isAll: true })"
                    />
                </el-tooltip>
                <el-tooltip :content="$t('commons.button.copy')">
                    <el-button class="-mr-3" link icon="CopyDocument" @click="copyText(tmpTags.join('\n'))" />
                </el-tooltip>
                <el-tooltip :content="$t('commons.button.clean')">
                    <el-button link icon="Close" @click="handleClean" />
                </el-tooltip>
            </template>
        </el-input-tag>
        <span v-if="props.egHelp" class="input-help">{{ props.egHelp }}</span>

        <FileList ref="fileRef" @choose="loadFile" />
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { copyText } from '@/utils/util';
import FileList from '@/components/file-list/index.vue';
const em = defineEmits(['update:tags']);

const tmpTags = ref([]);
const fileRef = ref();

const props = defineProps({
    egHelp: { type: String, default: 'key=val' },
    tags: { type: Array<string>, default: [] },

    withFile: { type: Boolean, default: false },
    baseDir: { type: String, default: '/' },
});
watch(
    () => props.tags,
    (newVal) => {
        tmpTags.value = newVal || [];
    },
);

const loadFile = async (path: string) => {
    handleAdd(path);
};

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
    handleUpdate();
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
