<template>
    <div>
        <el-drawer
            v-model="changeVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            width="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('database.permission')" :resource="form.name" :back="handleClose" />
            </template>
            <el-form v-loading="loading" :model="form" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('database.userBind')">
                            <el-tag>
                                {{ form.username }}
                            </el-tag>
                        </el-form-item>
                        <el-form-item :label="$t('database.permission')" prop="superUser">
                            <el-checkbox v-model="form.superUser">{{ $t('database.pgSuperUser') }}</el-checkbox>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="changeVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit()">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { changePrivileges } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const changeVisible = ref(false);
const form = reactive({
    database: '',
    name: '',
    username: '',
    superUser: true,
});

interface DialogProps {
    database: string;
    name: string;
    username: string;
    superUser: boolean;
}
const acceptParams = (params: DialogProps): void => {
    form.database = params.database;
    form.name = params.name;
    form.username = params.username;
    form.superUser = params.superUser;
    changeVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    changeVisible.value = false;
};

const onSubmit = async () => {
    let param = {
        name: form.name,
        database: form.database,
        username: form.username,
        superUser: form.superUser,
    };
    loading.value = true;
    await changePrivileges(param)
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
