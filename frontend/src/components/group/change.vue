<template>
    <div v-loading="loading">
        <DrawerPro v-model="drawerVisible" :header="$t('terminal.groupChange')" :back="handleClose" size="small">
            <el-form @submit.prevent ref="hostInfoRef" label-position="top" :model="dialogData" :rules="rules">
                <el-form-item :label="$t('commons.table.group')" prop="group">
                    <el-select filterable v-model="dialogData.groupID" clearable style="width: 100%">
                        <div v-for="item in groupList" :key="item.id">
                            <el-option :label="item.name" :value="item.id" />
                        </div>
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(hostInfoRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DrawerPro>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { getGroupList } from '@/api/modules/group';

const loading = ref();
interface DialogProps {
    group: string;
    groupType: string;
}
const drawerVisible = ref(false);
const dialogData = ref({
    groupID: 0,
    groupType: '',
});

const groupList = ref();
const acceptParams = (params: DialogProps): void => {
    dialogData.value.groupType = params.groupType;
    loadGroups(params.group);
    drawerVisible.value = true;
};
const emit = defineEmits(['search', 'change']);

const handleClose = () => {
    drawerVisible.value = false;
};

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    groupID: [Rules.requiredSelect],
});

const loadGroups = async (groupName: string) => {
    const res = await getGroupList(dialogData.value.groupType);
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
        emit('change', Number(dialogData.value.groupID));
        loading.value = false;
        drawerVisible.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
