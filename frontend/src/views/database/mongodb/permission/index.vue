<template>
    <div>
        <DrawerPro
            v-model="changeVisible"
            :header="$t('database.permission')"
            :resource="form.name"
            @close="handleClose"
            size="small"
        >
            <el-form v-loading="loading" :model="form" label-position="top">
                <el-form-item :label="$t('database.userBind')">
                    <el-tag>
                        {{ form.username }}
                    </el-tag>
                </el-form-item>
                <el-form-item :label="$t('database.permission')" prop="permission">
                    <el-select v-model="form.permission" class="w-full">
                        <el-option
                            v-for="item in permissionOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                        />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button :disabled="loading" @click="changeVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading || !form.permission" type="primary" @click="onSubmit()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DrawerPro>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { changeMongodbPrivileges, loadMongodbPrivileges } from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const changeVisible = ref(false);
const form = reactive({
    database: '',
    name: '',
    username: '',
    permission: '',
});

const permissionOptions = [
    { label: i18n.global.t('database.mongodbPermissionDbOwner'), value: 'dbOwner' },
    { label: i18n.global.t('database.mongodbPermissionRead'), value: 'read' },
    { label: i18n.global.t('database.mongodbPermissionReadWrite'), value: 'readWrite' },
    { label: i18n.global.t('database.mongodbPermissionUserAdmin'), value: 'userAdmin' },
];

interface DialogProps {
    database: string;
    name: string;
    username: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    form.database = params.database;
    form.name = params.name;
    form.username = params.username;
    form.permission = '';
    changeVisible.value = true;
    loading.value = true;
    await loadMongodbPrivileges({
        database: form.database,
        name: form.name,
        username: form.username,
    })
        .then((res) => {
            form.permission = res.data || '';
        })
        .finally(() => {
            loading.value = false;
        });
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    changeVisible.value = false;
};

const onSubmit = async () => {
    const param = {
        database: form.database,
        name: form.name,
        username: form.username,
        permission: form.permission,
    };
    loading.value = true;
    await changeMongodbPrivileges(param)
        .then(() => {
            loading.value = false;
            emit('search');
            changeVisible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
