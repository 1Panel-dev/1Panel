<template>
    <div>
        <DrawerPro
            v-model="bindVisible"
            :header="$t('database.userBind')"
            :resource="form.name"
            @close="handleClose"
            size="small"
        >
            <el-form v-loading="loading" ref="changeFormRef" :model="form" :rules="rules" label-position="top">
                <el-form-item :label="$t('commons.login.username')" prop="username">
                    <el-input v-model="form.username" />
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="password">
                    <el-input type="password" clearable show-password v-model="form.password" />
                    <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="bindVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(changeFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DrawerPro>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { bindMongodbUser } from '@/api/modules/database';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const bindVisible = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const changeFormRef = ref<FormInstance>();
const form = reactive({
    database: '',
    name: '',
    username: '',
    password: '',
});

const rules = reactive({
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
});

interface DialogProps {
    database: string;
    name: string;
}
const acceptParams = (params: DialogProps): void => {
    form.database = params.database;
    form.name = params.name;
    form.username = '';
    form.password = '';
    bindVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    bindVisible.value = false;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await bindMongodbUser(form)
            .then(() => {
                loading.value = false;
                emit('search');
                bindVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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
