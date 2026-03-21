<template>
    <div class="p-4">
        <el-alert v-if="!props.domain.dnsAutoGen" type="info" :closable="false" class="mb-4">
            DNS records were not auto-configured. Add the following records manually to your DNS zone.
        </el-alert>

        <el-card class="mb-4">
            <template #header>
                <span>MX Record</span>
            </template>
            <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="Name">{{ props.domain.name }}</el-descriptions-item>
                <el-descriptions-item label="Type">MX</el-descriptions-item>
                <el-descriptions-item label="Priority">10</el-descriptions-item>
                <el-descriptions-item label="Value">
                    mail.{{ props.domain.name }}
                    <CopyButton :content="'mail.' + props.domain.name" class="ml-2" />
                </el-descriptions-item>
            </el-descriptions>
        </el-card>

        <el-card class="mb-4">
            <template #header>
                <span>{{ $t('mail.spfRecord') }}</span>
            </template>
            <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="Name">{{ props.domain.name }}</el-descriptions-item>
                <el-descriptions-item label="Type">TXT</el-descriptions-item>
                <el-descriptions-item label="Value">
                    v=spf1 mx a ~all
                    <CopyButton content="v=spf1 mx a ~all" class="ml-2" />
                </el-descriptions-item>
            </el-descriptions>
        </el-card>

        <el-card class="mb-4">
            <template #header>
                <span>{{ $t('mail.dkimKey') }}</span>
            </template>
            <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="Name">mail._domainkey.{{ props.domain.name }}</el-descriptions-item>
                <el-descriptions-item label="Type">TXT</el-descriptions-item>
                <el-descriptions-item label="Value">
                    <div v-if="props.domain.dkimKey" style="word-break: break-all; max-width: 500px">
                        {{ props.domain.dkimKey }}
                        <CopyButton :content="props.domain.dkimKey" class="ml-2" />
                    </div>
                    <el-tag v-else type="warning" size="small">Not generated yet</el-tag>
                </el-descriptions-item>
            </el-descriptions>
        </el-card>

        <el-card>
            <template #header>
                <span>{{ $t('mail.dmarcRecord') }}</span>
            </template>
            <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="Name">_dmarc.{{ props.domain.name }}</el-descriptions-item>
                <el-descriptions-item label="Type">TXT</el-descriptions-item>
                <el-descriptions-item label="Value">
                    v=DMARC1; p=quarantine; rua=mailto:postmaster@{{ props.domain.name }}
                    <CopyButton :content="'v=DMARC1; p=quarantine; rua=mailto:postmaster@' + props.domain.name" class="ml-2" />
                </el-descriptions-item>
            </el-descriptions>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { Mail } from '@/api/interface/mail';

const props = defineProps<{ domain: Mail.DomainRes }>();
</script>
