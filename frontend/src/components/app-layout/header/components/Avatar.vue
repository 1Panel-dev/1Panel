<template>
    <el-dropdown trigger="click">
        <div class="avatar">
            <img src="@/assets/images/avatar.gif" alt="avatar" />
        </div>
        <template #dropdown>
            <el-dropdown-menu>
                <el-dropdown-item @click="openDialog('infoRef')">{{ $t('header.personalData') }}</el-dropdown-item>
                <el-dropdown-item @click="openDialog('passwordRef')">{{
                    $t('header.changePassword')
                }}</el-dropdown-item>
                <el-dropdown-item @click="logout" divided>{{ $t('header.logout') }}</el-dropdown-item>
            </el-dropdown-menu>
        </template>
    </el-dropdown>
    <InfoDialog ref="infoRef"></InfoDialog>
    <PasswordDialog ref="passwordRef"></PasswordDialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import InfoDialog from './info-dialog.vue';
import PasswordDialog from './password-dialog.vue';
import { ElMessageBox, ElMessage } from 'element-plus';
import { useRouter } from 'vue-router';
import { GlobalStore } from '@/store';
import { logOutApi } from '@/api/modules/login';
import i18n from '@/lang';

const router = useRouter();
const globalStore = GlobalStore();

const logout = () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.sureLogOut'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    }).then(() => {
        systemLogOut();
        router.push({ name: 'login' });
        globalStore.setLogStatus(false);
        ElMessage({
            type: 'success',
            message: i18n.global.t('commons.msg.operationSuccess'),
        });
    });
};

const systemLogOut = async () => {
    await logOutApi();
};

interface DialogExpose {
    openDialog: () => void;
}
const infoRef = ref<null | DialogExpose>(null);
const passwordRef = ref<null | DialogExpose>(null);

const openDialog = (refName: string) => {
    if (refName == 'infoRef') return infoRef.value?.openDialog();
    passwordRef.value?.openDialog();
};
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
