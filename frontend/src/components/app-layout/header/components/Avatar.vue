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
    <!-- infoDialog -->
    <InfoDialog ref="infoRef"></InfoDialog>
    <!-- passwordDialog -->
    <PasswordDialog ref="passwordRef"></PasswordDialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import InfoDialog from './info-dialog.vue';
import PasswordDialog from './password-dialog.vue';
import { ElMessageBox, ElMessage } from 'element-plus';
import { useRouter } from 'vue-router';
import { GlobalStore } from '@/store';

const router = useRouter();
const globalStore = GlobalStore();

// 退出登录
const logout = () => {
    ElMessageBox.confirm('您是否确认退出登录?', '温馨提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
    }).then(() => {
        router.push({ name: 'login' });
        globalStore.setUserInfo('');
        ElMessage({
            type: 'success',
            message: '退出登录成功！',
        });
    });
};

interface DialogExpose {
    openDialog: () => void;
}
const infoRef = ref<null | DialogExpose>(null);
const passwordRef = ref<null | DialogExpose>(null);
// 打开修改密码和个人信息弹窗
const openDialog = (refName: string) => {
    if (refName == 'infoRef') return infoRef.value?.openDialog();
    passwordRef.value?.openDialog();
};
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
