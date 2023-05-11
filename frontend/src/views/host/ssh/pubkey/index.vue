<template>
    <div>
        <el-drawer
            v-model="drawerVisiable"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('ssh.pubkey')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('ssh.encryptionMode')" prop="encryptionMode">
                            <el-select v-model="form.encryptionMode">
                                <el-option label="ED25519" value="ed25519" />
                                <el-option label="ECDSA" value="ecdsa" />
                                <el-option label="RSA" value="rsa" />
                                <el-option label="DSA" value="dsa" />
                            </el-select>

                            <el-button link @click="onDownload" type="primary" class="margintop">
                                {{ form.primaryKey ? $t('ssh.reGenerate') : $t('ssh.generate') }}
                            </el-button>
                        </el-form-item>
                        <el-form-item :label="$t('ssh.key')" prop="primaryKey" v-if="form.encryptionMode">
                            <el-input
                                v-model="form.primaryKey"
                                :autosize="{ minRows: 5, maxRows: 15 }"
                                type="textarea"
                            />
                            <div v-if="form.primaryKey">
                                <el-button link type="primary" icon="CopyDocument" class="margintop" @click="loadSSLs">
                                    {{ $t('file.copy') }}
                                </el-button>
                                <el-button link type="primary" icon="Download" class="margintop" @click="onDownload">
                                    {{ $t('commons.button.download') }}
                                </el-button>
                            </div>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';

const loading = ref();
const drawerVisiable = ref();

const form = reactive({
    encryptionMode: '',
    primaryKey: '',
});

const acceptParams = async (): Promise<void> => {
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const loadSSLs = async () => {};

const onDownload = async () => {};

const handleClose = () => {
    emit('search');
    drawerVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.margintop {
    margin-top: 10px;
}
</style>
