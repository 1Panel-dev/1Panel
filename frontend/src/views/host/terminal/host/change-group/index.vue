<template>
    <div v-loading="loading">
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('terminal.groupChange')" :back="handleClose" />
            </template>
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form ref="hostInfoRef" label-position="top" :model="dialogData" :rules="rules">
                        <el-form-item :label="$t('commons.table.group')" prop="group">
                            <el-select filterable v-model="dialogData.group" clearable style="width: 100%">
                                <el-option
                                    v-for="item in groupList"
                                    :key="item.id"
                                    :label="item.name"
                                    :value="item.name"
                                />
                            </el-select>
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="onSubmit(hostInfoRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { editHostGroup, getGroupList } from '@/api/modules/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
interface DialogProps {
    id: number;
    group: string;
}
const drawerVisiable = ref(false);
const dialogData = ref<DialogProps>({
    id: 0,
    group: '',
});

const groupList = ref();
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    drawerVisiable.value = true;
    loadGroups();
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    group: [Rules.requiredSelect],
});

const loadGroups = async () => {
    const res = await getGroupList({ type: 'host' });
    groupList.value = res.data;
};

const onSubmit = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await editHostGroup(dialogData.value)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                drawerVisiable.value = false;
                emit('search');
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
