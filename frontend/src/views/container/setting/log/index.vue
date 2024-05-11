<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            @close="handleClose"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('container.cutLog')" :back="handleClose" />
            </template>
            <el-alert class="common-prompt" :closable="false" type="warning">
                <template #default>
                    <ul style="margin-left: -20px">
                        <li>{{ $t('container.cutLogHelper1') }}</li>
                        <li>{{ $t('container.cutLogHelper2') }}</li>
                        <li>{{ $t('container.cutLogHelper3') }}</li>
                    </ul>
                </template>
            </el-alert>
            <el-form :model="form" ref="formRef" :rules="rules" v-loading="loading" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item prop="logMaxSize" :label="$t('container.maxSize')">
                            <el-input v-model.number="form.logMaxSize">
                                <template #append>
                                    <el-select v-model="form.sizeUnit" style="width: 70px">
                                        <el-option label="Byte" value="b"></el-option>
                                        <el-option label="KB" value="k"></el-option>
                                        <el-option label="MB" value="m"></el-option>
                                        <el-option label="GB" value="g"></el-option>
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item prop="logMaxFile" :label="$t('container.maxFile')">
                            <el-input v-model.number="form.logMaxFile" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitSave"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { updateLogOption } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';

const loading = ref();
const drawerVisible = ref();
const confirmDialogRef = ref();
const formRef = ref();

interface DialogProps {
    logMaxSize: string;
    logMaxFile: number;
}

const form = reactive({
    logMaxSize: 10,
    logMaxFile: 3,
    sizeUnit: 'm',
});
const rules = reactive({
    logMaxSize: [checkNumberRange(1, 1024000), Rules.number],
    logMaxFile: [checkNumberRange(1, 100), Rules.number],
});

const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = (params: DialogProps): void => {
    form.logMaxFile = params.logMaxFile || 3;
    if (params.logMaxSize) {
        form.logMaxSize = loadSize(params.logMaxSize);
    } else {
        form.logMaxSize = 10;
        form.sizeUnit = 'm';
    }
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

const onSubmitSave = async () => {
    loading.value = true;
    await updateLogOption(form.logMaxSize + form.sizeUnit, form.logMaxFile + '')
        .then(() => {
            loading.value = false;
            drawerVisible.value = false;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadSize = (value: string) => {
    if (value.indexOf('k') !== -1 || value.indexOf('KB') !== -1) {
        form.sizeUnit = 'k';
        return Number(value.replaceAll('k', '').replaceAll('KB', ''));
    }
    if (value.indexOf('m') !== -1 || value.indexOf('MB') !== -1) {
        form.sizeUnit = 'm';
        return Number(value.replaceAll('m', '').replaceAll('MB', ''));
    }
    if (value.indexOf('g') !== -1 || value.indexOf('GB') !== -1) {
        form.sizeUnit = 'g';
        return Number(value.replaceAll('g', '').replaceAll('GB', ''));
    }
    if (value.indexOf('b') !== -1 || value.indexOf('B') !== -1) {
        form.sizeUnit = 'b';
        return Number(value.replaceAll('b', '').replaceAll('B', ''));
    }
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
