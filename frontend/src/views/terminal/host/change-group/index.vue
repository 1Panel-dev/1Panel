<template>
    <DrawerPro v-model="drawerVisible" :header="$t('terminal.groupChange')" :back="handleClose" size="small">
        <el-form
            @submit.prevent
            ref="hostInfoRef"
            label-position="top"
            :model="dialogData"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('commons.table.group')" prop="group">
                <el-select filterable v-model="dialogData.groupID" clearable class="w-full">
                    <div v-for="item in groupList" :key="item.id">
                        <el-option :label="item.name" :value="item.id" />
                    </div>
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(hostInfoRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { editHostGroup } from '@/api/modules/terminal';
import { getGroupList } from '@/api/modules/group';
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
    const res = await getGroupList('host');
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
