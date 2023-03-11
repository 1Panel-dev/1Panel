<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('container.createVolume')" :back="handleClose" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-form
                    ref="formRef"
                    v-loading="loading"
                    label-position="top"
                    :model="form"
                    :rules="rules"
                    label-width="80px"
                >
                    <el-form-item :label="$t('container.volumeName')" prop="name">
                        <el-input clearable v-model.trim="form.name" />
                    </el-form-item>
                    <el-form-item :label="$t('container.driver')" prop="driver">
                        <el-select v-model="form.driver">
                            <el-option label="local" value="local" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.option')" prop="optionStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :autosize="{ minRows: 2, maxRows: 4 }"
                            v-model="form.optionStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')" prop="labelStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :autosize="{ minRows: 2, maxRows: 4 }"
                            v-model="form.labelStr"
                        />
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { createVolume } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const drawerVisiable = ref(false);
const form = reactive({
    name: '',
    driver: 'local',
    labelStr: '',
    labels: [] as Array<string>,
    optionStr: '',
    options: [] as Array<string>,
});

const acceptParams = (): void => {
    form.name = '';
    form.labels = [];
    form.labelStr = '';
    form.options = [];
    form.optionStr = '';
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput, Rules.volumeName],
    driver: [Rules.requiredSelect],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.labelStr !== '') {
            form.labels = form.labelStr.split('\n');
        }
        if (form.optionStr !== '') {
            form.options = form.optionStr.split('\n');
        }
        loading.value = true;
        await createVolume(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisiable.value = false;
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
