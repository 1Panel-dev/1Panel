<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('toolbox.fail2ban.banAction')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('toolbox.fail2ban.banAction')"
                            prop="banAction"
                            :rules="Rules.requiredSelect"
                        >
                            <el-select v-model="form.banAction">
                                <el-option value="iptables-allports" class="option">
                                    <span class="option-content">iptables-allports</span>
                                    <span class="input-help option-helper">
                                        {{
                                            $t('toolbox.fail2ban.banActionOption', ['iptables']) +
                                            $t('toolbox.fail2ban.allPorts')
                                        }}
                                    </span>
                                </el-option>
                                <el-option value="iptables-multiport" class="option">
                                    <span class="option-content">iptables-multiport</span>
                                    <span class="input-help option-helper">
                                        {{ $t('toolbox.fail2ban.banActionOption', ['iptables']) }}
                                    </span>
                                </el-option>
                                <el-option value="firewallcmd-ipset" class="option">
                                    <span class="option-content">firewallcmd-ipset</span>
                                    <span class="input-help option-helper">
                                        {{ $t('toolbox.fail2ban.banActionOption', ['firewallcmd ipset']) }}
                                    </span>
                                </el-option>
                                <el-option value="ufw" class="option">
                                    <span class="option-content">ufw</span>
                                    <span class="input-help option-helper">
                                        {{ $t('toolbox.fail2ban.banActionOption', ['ufw']) }}
                                    </span>
                                </el-option>
                            </el-select>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
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
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateFail2ban } from '@/api/modules/toolbox';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    banAction: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    banAction: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.banAction = params.banAction;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('toolbox.fail2ban.banAction'), form.banAction]),
            i18n.global.t('toolbox.fail2ban.fail2banChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            await updateFail2ban({ key: 'banaction', value: form.banAction })
                .then(async () => {
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    loading.value = false;
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.option {
    height: 50px;
}
.option-content {
    float: left;
    position: relative;
    display: block;
}
.option-helper {
    float: left;
    display: block;
    position: absolute;
    margin-top: 20px;
}
</style>
