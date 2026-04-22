<template>
    <el-form-item prop="advanced">
        <el-checkbox v-model="form.advanced" :label="$t('app.advanced')" size="large" />
    </el-form-item>
    <div v-if="form.advanced">
        <el-form-item :label="$t('app.containerName')" prop="containerName">
            <el-input v-model.trim="form.containerName" :placeholder="$t('app.containerNameHelper')" />
        </el-form-item>
        <el-form-item v-if="showAllowPort" prop="allowPort">
            <el-checkbox v-model="form.allowPort" :label="$t('app.allowPort')" size="large" />
            <span class="input-help">{{ $t('app.allowPortHelper') }}</span>
        </el-form-item>
        <el-form-item v-if="showSpecifyIP && form.allowPort" :label="$t('app.specifyIP')" prop="specifyIP">
            <el-input v-model="form.specifyIP" />
            <span class="input-help">{{ $t('app.specifyIPHelper') }}</span>
        </el-form-item>
        <el-form-item v-if="showRestartPolicy" :label="$t('container.restartPolicy')" prop="restartPolicy">
            <el-select v-model="form.restartPolicy" class="p-w-300">
                <el-option :label="$t('container.no')" value="no" />
                <el-option :label="$t('container.always')" value="always" />
                <el-option :label="$t('container.onFailure')" value="on-failure" />
                <el-option :label="$t('container.unlessStopped')" value="unless-stopped" />
            </el-select>
        </el-form-item>
        <el-form-item :label="$t('container.cpuQuota')" prop="cpuQuota" :rules="checkNumberRange(0, limits.cpu)">
            <el-input type="number" class="!w-2/5" v-model.number="form.cpuQuota" maxlength="5">
                <template #append>{{ $t('app.cpuCore') }}</template>
            </el-input>
            <span class="input-help">
                {{ $t('container.limitHelper', [limits.cpu]) }}{{ $t('commons.units.core') }}
            </span>
        </el-form-item>
        <el-form-item
            :label="$t('container.memoryLimit')"
            prop="memoryLimit"
            :rules="checkNumberRange(0, limits.memory)"
        >
            <el-input class="!w-2/5" v-model.number="form.memoryLimit" maxlength="10">
                <template #append>
                    <el-select v-model="form.memoryUnit" class="p-w-100" @change="changeUnit">
                        <el-option label="MB" value="M" />
                        <el-option label="GB" value="G" />
                    </el-select>
                </template>
            </el-input>
            <span class="input-help">{{ $t('container.limitHelper', [limits.memory]) }}{{ form.memoryUnit }}B</span>
        </el-form-item>
        <el-form-item v-if="showPullImage" prop="pullImage">
            <el-checkbox v-model="form.pullImage" :label="$t('app.pullImage')" />
            <span class="input-help">{{ $t('app.pullImageHelper') }}</span>
        </el-form-item>
        <el-form-item v-if="showCompose" prop="editCompose">
            <el-checkbox v-model="form.editCompose" :label="$t('app.editCompose')" />
            <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
        </el-form-item>
        <div v-if="showCompose && form.editCompose">
            <CodemirrorPro v-model="form.dockerCompose" mode="yaml" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref, toRef } from 'vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { checkNumberRange } from '@/global/form-rules';
import { loadResourceLimit } from '@/api/modules/container';
import { Container } from '@/api/interface/container';

const props = withDefaults(
    defineProps<{
        form: {
            advanced: boolean;
            containerName: string;
            allowPort: boolean;
            specifyIP: string;
            restartPolicy: string;
            cpuQuota: number;
            memoryLimit: number;
            memoryUnit: string;
            pullImage: boolean;
            editCompose: boolean;
            dockerCompose: string;
        };
        showAllowPort?: boolean;
        showSpecifyIP?: boolean;
        showRestartPolicy?: boolean;
        showPullImage?: boolean;
        showCompose?: boolean;
        autoLoadLimit?: boolean;
    }>(),
    {
        showAllowPort: true,
        showSpecifyIP: true,
        showRestartPolicy: true,
        showPullImage: true,
        showCompose: true,
        autoLoadLimit: true,
    },
);

const form = toRef(props, 'form');
const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});
const oldMemory = ref(0);

const changeUnit = () => {
    if (form.value.memoryUnit === 'M') {
        limits.value.memory = oldMemory.value;
    } else {
        limits.value.memory = Number((oldMemory.value / 1024).toFixed(2));
    }
};

const loadLimit = async () => {
    const res = await loadResourceLimit();
    limits.value = res.data;
    limits.value.memory = Number((limits.value.memory / 1024 / 1024).toFixed(2));
    oldMemory.value = limits.value.memory;
};

onMounted(() => {
    if (props.autoLoadLimit) {
        loadLimit();
    }
});
</script>
