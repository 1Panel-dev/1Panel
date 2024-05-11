<template>
    <div v-loading="loading">
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('terminal.groupChange')" :back="handleClose" />
            </template>
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form @submit.prevent ref="hostInfoRef" label-position="top" :model="dialogData" :rules="rules">
                        <el-form-item :label="$t('commons.table.group')" prop="group">
                            <el-select filterable v-model="dialogData.groupID" clearable style="width: 100%">
                                <div v-for="item in groupList" :key="item.id">
                                    <el-option :label="item.name" :value="item.id" />
                                </div>
                            </el-select>
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { editHostGroup } from '@/api/modules/host';
import { GetGroupList } from '@/api/modules/group';
import DrawerHeader from '@/components/drawer-header/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
interface DialogProps {
    id: number;
    group: string;
}
const drawerVisible = ref(false);
const dialogData = ref({
    id: 0,
    groupID: 0,
});

const groupList = ref();
const acceptParams = (params: DialogProps): void => {
    dialogData.value.id = params.id;
    loadGroups(params.group);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    groupID: [Rules.requiredSelect],
});

const loadGroups = async (groupName: string) => {
    const res = await GetGroupList({ type: 'host' });
    groupList.value = res.data;
    for (const group of groupList.value) {
        if (group.name === groupName) {
            dialogData.value.groupID = group.id;
            break;
        }
    }
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
                drawerVisible.value = false;
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
