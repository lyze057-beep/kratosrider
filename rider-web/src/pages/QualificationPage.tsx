import React, { useState, useEffect } from 'react';
import { Card, Button, Steps, Form, Input, Upload, message, Descriptions, Tag } from 'antd';
import { UploadOutlined, CheckCircleOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { qualificationApi, QualificationInfo } from '../api/qualification';
import { useAuthStore } from '../store/authStore';

const { Step } = Steps;

const STATUS_TEXT: Record<number, string> = {
  0: '未提交',
  1: '待审核',
  2: '已通过',
  3: '已拒绝'
};

const STATUS_COLOR: Record<number, string> = {
  0: 'default',
  1: 'processing',
  2: 'success',
  3: 'error'
};

const QualificationPage: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [qualification, setQualification] = useState<QualificationInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [activeStep, setActiveStep] = useState(0);
  const [idCardForm] = Form.useForm();
  const [licenseForm] = Form.useForm();
  const [healthForm] = Form.useForm();

  useEffect(() => {
    loadQualification();
  }, []);

  const loadQualification = async () => {
    try {
      const result = await qualificationApi.getVerificationDetail(Number(userInfo?.userId || 0));
      setQualification(result.qualification);
      // 根据认证状态设置当前步骤
      if (result.qualification?.id_card_status === 2) {
        if (result.qualification?.driver_license_status === 2) {
          if (result.qualification?.health_certificate_status === 2) {
            setActiveStep(3);
          } else {
            setActiveStep(2);
          }
        } else {
          setActiveStep(1);
        }
      }
    } catch (error) {
      console.error('Load qualification error:', error);
    }
  };

  const handleSubmitIDCard = async (values: any) => {
    setLoading(true);
    try {
      await qualificationApi.submitIDCard({
        rider_id: Number(userInfo?.userId || 0),
        real_name: values.real_name,
        id_card_number: values.id_card_number,
        id_card_front: values.id_card_front?.file?.response?.url || '',
        id_card_back: values.id_card_back?.file?.response?.url || ''
      });
      message.success('身份证认证提交成功');
      loadQualification();
      setActiveStep(1);
    } catch (error) {
      message.error('提交失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitLicense = async (values: any) => {
    setLoading(true);
    try {
      await qualificationApi.submitDriverLicense({
        rider_id: Number(userInfo?.userId || 0),
        driver_license_number: values.driver_license_number,
        driver_license_image: values.driver_license_image?.file?.response?.url || ''
      });
      message.success('驾驶证认证提交成功');
      loadQualification();
      setActiveStep(2);
    } catch (error) {
      message.error('提交失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitHealth = async (values: any) => {
    setLoading(true);
    try {
      await qualificationApi.submitHealthCertificate({
        rider_id: Number(userInfo?.userId || 0),
        health_certificate_number: values.health_certificate_number,
        health_certificate_image: values.health_certificate_image?.file?.response?.url || ''
      });
      message.success('健康证认证提交成功');
      loadQualification();
      setActiveStep(3);
    } catch (error) {
      message.error('提交失败');
    } finally {
      setLoading(false);
    }
  };

  const uploadProps = {
    name: 'file',
    action: '/api/upload',
    headers: {
      authorization: 'authorization-text',
    },
    onChange(info: any) {
      if (info.file.status === 'done') {
        message.success(`${info.file.name} 上传成功`);
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} 上传失败`);
      }
    },
  };

  return (
    <div style={{ padding: 24 }}>
      <Card title="资质认证" style={{ marginBottom: 16 }}>
        <Steps current={activeStep} style={{ marginBottom: 32 }}>
          <Step title="身份证认证" icon={qualification?.id_card_status === 2 ? <CheckCircleOutlined /> : <ClockCircleOutlined />} />
          <Step title="驾驶证认证" icon={qualification?.driver_license_status === 2 ? <CheckCircleOutlined /> : <ClockCircleOutlined />} />
          <Step title="健康证认证" icon={qualification?.health_certificate_status === 2 ? <CheckCircleOutlined /> : <ClockCircleOutlined />} />
        </Steps>

        {qualification && (
          <Descriptions title="当前认证状态" bordered style={{ marginBottom: 24 }}>
            <Descriptions.Item label="身份证">
              <Tag color={STATUS_COLOR[qualification.id_card_status]}>
                {STATUS_TEXT[qualification.id_card_status]}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="驾驶证">
              <Tag color={STATUS_COLOR[qualification.driver_license_status]}>
                {STATUS_TEXT[qualification.driver_license_status]}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="健康证">
              <Tag color={STATUS_COLOR[qualification.health_certificate_status]}>
                {STATUS_TEXT[qualification.health_certificate_status]}
              </Tag>
            </Descriptions.Item>
          </Descriptions>
        )}

        {activeStep === 0 && qualification?.id_card_status !== 2 && (
          <Card title="身份证认证" type="inner">
            <Form form={idCardForm} onFinish={handleSubmitIDCard} layout="vertical">
              <Form.Item
                name="real_name"
                label="真实姓名"
                rules={[{ required: true, message: '请输入真实姓名' }]}
              >
                <Input placeholder="请输入真实姓名" />
              </Form.Item>
              <Form.Item
                name="id_card_number"
                label="身份证号"
                rules={[{ required: true, message: '请输入身份证号' }]}
              >
                <Input placeholder="请输入身份证号" />
              </Form.Item>
              <Form.Item
                name="id_card_front"
                label="身份证正面"
                rules={[{ required: true, message: '请上传身份证正面' }]}
              >
                <Upload {...uploadProps}>
                  <Button icon={<UploadOutlined />}>上传身份证正面</Button>
                </Upload>
              </Form.Item>
              <Form.Item
                name="id_card_back"
                label="身份证反面"
                rules={[{ required: true, message: '请上传身份证反面' }]}
              >
                <Upload {...uploadProps}>
                  <Button icon={<UploadOutlined />}>上传身份证反面</Button>
                </Upload>
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  提交认证
                </Button>
              </Form.Item>
            </Form>
          </Card>
        )}

        {activeStep === 1 && qualification?.driver_license_status !== 2 && (
          <Card title="驾驶证认证" type="inner">
            <Form form={licenseForm} onFinish={handleSubmitLicense} layout="vertical">
              <Form.Item
                name="driver_license_number"
                label="驾驶证号"
                rules={[{ required: true, message: '请输入驾驶证号' }]}
              >
                <Input placeholder="请输入驾驶证号" />
              </Form.Item>
              <Form.Item
                name="driver_license_image"
                label="驾驶证照片"
                rules={[{ required: true, message: '请上传驾驶证照片' }]}
              >
                <Upload {...uploadProps}>
                  <Button icon={<UploadOutlined />}>上传驾驶证照片</Button>
                </Upload>
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  提交认证
                </Button>
              </Form.Item>
            </Form>
          </Card>
        )}

        {activeStep === 2 && qualification?.health_certificate_status !== 2 && (
          <Card title="健康证认证" type="inner">
            <Form form={healthForm} onFinish={handleSubmitHealth} layout="vertical">
              <Form.Item
                name="health_certificate_number"
                label="健康证号"
                rules={[{ required: true, message: '请输入健康证号' }]}
              >
                <Input placeholder="请输入健康证号" />
              </Form.Item>
              <Form.Item
                name="health_certificate_image"
                label="健康证照片"
                rules={[{ required: true, message: '请上传健康证照片' }]}
              >
                <Upload {...uploadProps}>
                  <Button icon={<UploadOutlined />}>上传健康证照片</Button>
                </Upload>
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  提交认证
                </Button>
              </Form.Item>
            </Form>
          </Card>
        )}

        {activeStep === 3 && (
          <Card type="inner">
            <div style={{ textAlign: 'center', padding: 40 }}>
              <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a' }} />
              <h3 style={{ marginTop: 16 }}>恭喜！所有资质认证已完成</h3>
              <p>您现在可以开始接单了</p>
            </div>
          </Card>
        )}
      </Card>
    </div>
  );
};

export default QualificationPage;
