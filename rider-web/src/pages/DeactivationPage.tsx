import React, { useState, useEffect } from 'react';
import { Card, Button, Radio, Input, message, Alert, Steps, Modal } from 'antd';
import { ExclamationCircleOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { deactivationApi, DeactivationInfo, DEACTIVATION_REASON, DEACTIVATION_REASON_TEXT, DEACTIVATION_STATUS } from '../api/deactivation';
import { useAuthStore } from '../store/authStore';

const { Step } = Steps;
const { confirm } = Modal;

const DeactivationPage: React.FC = () => {
  useAuthStore();
  const [deactivation, setDeactivation] = useState<DeactivationInfo | null>(null);
  const [reasonType, setReasonType] = useState<number>(DEACTIVATION_REASON.PERSONAL);
  const [reasonDetail, setReasonDetail] = useState('');
  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);

  useEffect(() => {
    loadDeactivationStatus();
  }, []);

  const loadDeactivationStatus = async () => {
    try {
      const result = await deactivationApi.getDeactivationStatus();
      if (result.deactivation) {
        setDeactivation(result.deactivation);
        // 根据状态设置步骤
        switch (result.deactivation.status) {
          case DEACTIVATION_STATUS.PENDING:
            setCurrentStep(1);
            break;
          case DEACTIVATION_STATUS.APPROVED:
            setCurrentStep(2);
            break;
          case DEACTIVATION_STATUS.COMPLETED:
            setCurrentStep(3);
            break;
          default:
            setCurrentStep(0);
        }
      }
    } catch (error) {
      console.error('Load deactivation status error:', error);
    }
  };

  const handleApply = () => {
    confirm({
      title: '确认申请注销账号?',
      icon: <ExclamationCircleOutlined />,
      content: '账号注销后将无法恢复，请谨慎操作！',
      okText: '确认注销',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        setLoading(true);
        try {
          await deactivationApi.applyDeactivation({
            reason_type: reasonType,
            reason_detail: reasonDetail
          });
          message.success('注销申请已提交');
          loadDeactivationStatus();
        } catch (error) {
          message.error('申请失败');
        } finally {
          setLoading(false);
        }
      }
    });
  };

  const handleCancel = async () => {
    confirm({
      title: '确认取消注销申请?',
      icon: <ExclamationCircleOutlined />,
      content: '取消后您可以继续使用账号',
      okText: '确认取消',
      cancelText: '保留申请',
      onOk: async () => {
        try {
          await deactivationApi.cancelDeactivation();
          message.success('注销申请已取消');
          setDeactivation(null);
          setCurrentStep(0);
        } catch (error) {
          message.error('取消失败');
        }
      }
    });
  };

  return (
    <div style={{ padding: 24 }}>
      <Card title="账号注销" style={{ maxWidth: 800, margin: '0 auto' }}>
        <Alert
          message="账号注销须知"
          description="账号注销后，您的所有数据将被清除且无法恢复。请确保已结清所有收入和订单。"
          type="warning"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Steps current={currentStep} style={{ marginBottom: 32 }}>
          <Step title="申请注销" icon={<ExclamationCircleOutlined />} />
          <Step title="等待审核" icon={<CheckCircleOutlined />} />
          <Step title="审核通过" icon={<CheckCircleOutlined />} />
          <Step title="注销完成" icon={<CloseCircleOutlined />} />
        </Steps>

        {currentStep === 0 && !deactivation && (
          <div>
            <h3>请选择注销原因</h3>
            <Radio.Group
              value={reasonType}
              onChange={(e) => setReasonType(e.target.value)}
              style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 24 }}
            >
              {Object.entries(DEACTIVATION_REASON_TEXT).map(([key, text]) => (
                <Radio key={key} value={Number(key)}>{text}</Radio>
              ))}
            </Radio.Group>

            <Input.TextArea
              placeholder="请详细说明注销原因（选填）"
              value={reasonDetail}
              onChange={(e) => setReasonDetail(e.target.value)}
              rows={4}
              style={{ marginBottom: 24 }}
            />

            <Button type="primary" danger onClick={handleApply} loading={loading} block>
              申请注销账号
            </Button>
          </div>
        )}

        {currentStep === 1 && deactivation?.status === DEACTIVATION_STATUS.PENDING && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <CheckCircleOutlined style={{ fontSize: 64, color: '#faad14' }} />
            <h3 style={{ marginTop: 16 }}>注销申请已提交</h3>
            <p>您的注销申请正在审核中，请耐心等待</p>
            <p>注销原因：{DEACTIVATION_REASON_TEXT[deactivation.reason_type]}</p>
            {deactivation.reason_detail && <p>详细原因：{deactivation.reason_detail}</p>}
            <p>申请时间：{deactivation.applied_at}</p>
            <Button onClick={handleCancel} style={{ marginTop: 16 }}>
              取消注销申请
            </Button>
          </div>
        )}

        {currentStep === 2 && deactivation?.status === DEACTIVATION_STATUS.APPROVED && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a' }} />
            <h3 style={{ marginTop: 16 }}>注销申请已通过</h3>
            <p>您的账号将在7天后自动注销</p>
            <p>在此期间您可以随时取消注销</p>
            <Button onClick={handleCancel} type="primary" style={{ marginTop: 16 }}>
              取消注销
            </Button>
          </div>
        )}

        {currentStep === 3 && deactivation?.status === DEACTIVATION_STATUS.COMPLETED && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <CloseCircleOutlined style={{ fontSize: 64, color: '#ff4d4f' }} />
            <h3 style={{ marginTop: 16 }}>账号已注销</h3>
            <p>您的账号已成功注销</p>
            <p>感谢您的使用，祝您生活愉快！</p>
          </div>
        )}
      </Card>
    </div>
  );
};

export default DeactivationPage;
