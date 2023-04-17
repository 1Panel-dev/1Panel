<template>
    <el-drawer v-model="deleteVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('container.imageDelete')" :back="handleClose" />
        </template>
        <el-form @submit.prevent :model="deleteForm" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('container.tag')" prop="tagName">
                        <el-checkbox-group v-model="deleteForm.deleteTags">
                            <div>
                                <el-checkbox
                                    style="width: 100%"
                                    v-for="item in deleteForm.tags"
                                    :key="item"
                                    :value="item"
                                    :label="item"
                                />
                            </div>
                        </el-checkbox-group>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="deleteVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="deleteForm.deleteTags.length === 0" @click="batchDelete()">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { ElForm } from 'element-plus';
import { imageRemove } from '@/api/modules/container';
import { useDeleteData } from '@/hooks/use-delete-data';
import DrawerHeader from '@/components/drawer-header/index.vue';

const deleteVisiable = ref(false);
const deleteForm = reactive({
    tags: [] as Array<string>,
    deleteTags: [] as Array<string>,
});

interface DialogProps {
    tags: Array<string>;
}
const acceptParams = (params: DialogProps) => {
    deleteVisiable.value = true;
    deleteForm.deleteTags = [];
    deleteForm.tags = params.tags;
};
const handleClose = () => {
    deleteVisiable.value = false;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const batchDelete = async () => {
    let names: Array<string> = [];
    for (const item of deleteForm.deleteTags) {
        names.push(item);
    }
    await useDeleteData(imageRemove, { names: names }, 'commons.msg.delete');
    deleteVisiable.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
