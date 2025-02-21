<template>
    <DialogPro v-model="visible" :title="$t('app.detail')">
        <div class="mt-5">
            <el-descriptions border :column="1">
                <el-descriptions-item v-for="(item, key) in list" :label="item.label" :key="key">
                    {{ item.value }}
                    <CopyButton v-if="!item.hideCopy" :content="item.value" type="icon" />
                </el-descriptions-item>
            </el-descriptions>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const list = ref();
const visible = ref(false);

interface DialogProps {
    list: Array<string>;
}

const acceptParams = (props: DialogProps): void => {
    visible.value = true;
    list.value = props.list;
};

defineExpose({
    acceptParams,
});
</script>
