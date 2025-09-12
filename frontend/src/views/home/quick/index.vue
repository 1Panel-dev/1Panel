<template>
    <DrawerPro v-model="drawerVisible" :header="$t('menu.home')" @close="handleClose">
        <div>
            <ComplexTable :heightDiff="1" :data="quickOptions" :show-header="false" row-key="id">
                <el-table-column prop="title" :label="$t('setting.menu')">
                    <template #default="{ row }">
                        {{ i18n.global.t(row.title) }}
                    </template>
                </el-table-column>
                <el-table-column prop="isShow" :label="$t('setting.ifShow')">
                    <template #default="{ row }">
                        <el-switch v-model="row.isShow" />
                        <div v-if="row.name === 'File' && row.isShow">
                            <el-input v-model="row.detail" class="w-full">
                                <template #prepend>
                                    <el-button
                                        v-if="row.name === 'File' && row.isShow"
                                        icon="Folder"
                                        @click="fileRef.acceptParams({ path: row.detail, isAll: true })"
                                    />
                                </template>
                            </el-input>
                        </div>
                    </template>
                </el-table-column>
            </ComplexTable>
        </div>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onChangeShow">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
        <FileList ref="fileRef" @choose="loadDir" />
    </DrawerPro>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import FileList from '@/components/file-list/index.vue';
import { changeQuick, loadQuickOption } from '@/api/modules/dashboard';

const emit = defineEmits<{ (e: 'search'): void }>();
const drawerVisible = ref();
const loading = ref();
const acceptParams = (): void => {
    search();
    drawerVisible.value = true;
};
const quickOptions = ref([]);
const fileRef = ref();

const search = async () => {
    loading.value = true;
    await loadQuickOption()
        .then((res) => {
            loading.value = false;
            quickOptions.value = res.data || [];
        })
        .catch(() => {
            loading.value = false;
        });
};
const onChangeShow = async () => {
    loading.value = true;
    await changeQuick(quickOptions.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            drawerVisible.value = false;
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadDir = async (path: string) => {
    for (const item of quickOptions.value) {
        if (item.name === 'File') {
            item.detail = path;
        }
    }
};

const handleClose = () => {
    drawerVisible.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
