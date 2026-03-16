import React, { useState, useEffect } from 'react';
import { Card, Button, message, Tabs, List, Tag, Progress, Statistic, Row, Col, QRCode } from 'antd';
import { ShareAltOutlined, GiftOutlined, UserAddOutlined, CopyOutlined } from '@ant-design/icons';
import { referralApi, InviteCodeInfo, InviteRecord, TaskInfo, RewardRecord } from '../api/referral';

const { TabPane } = Tabs;

const ReferralPage: React.FC = () => {
  const [inviteInfo, setInviteInfo] = useState<InviteCodeInfo | null>(null);
  const [inviteRecords, setInviteRecords] = useState<InviteRecord[]>([]);
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [rewards, setRewards] = useState<RewardRecord[]>([]);
  const [stats, setStats] = useState({
    totalInvited: 0,
    validInvited: 0,
    totalRewards: 0
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [infoRes, recordsRes, tasksRes, rewardsRes, statsRes] = await Promise.all([
        referralApi.getMyReferralInfo(),
        referralApi.getInviteRecordList(),
        referralApi.getTaskList(),
        referralApi.getRewardRecordList(),
        referralApi.getReferralStatistics()
      ]);

      setInviteInfo(infoRes.invite_info);
      setInviteRecords(recordsRes.records || []);
      setTasks(tasksRes.tasks || []);
      setRewards(rewardsRes.records || []);
      setStats({
        totalInvited: statsRes.total_invited || 0,
        validInvited: statsRes.valid_invited || 0,
        totalRewards: statsRes.total_rewards || 0
      });
    } catch (error) {
      console.error('Load referral data error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateCode = async () => {
    try {
      const result = await referralApi.generateInviteCode();
      setInviteInfo(result.invite_info);
      message.success('邀请码生成成功');
    } catch (error) {
      message.error('生成失败');
    }
  };

  const handleCopyCode = () => {
    if (inviteInfo?.invite_code) {
      navigator.clipboard.writeText(inviteInfo.invite_code);
      message.success('邀请码已复制');
    }
  };

  const handleCopyLink = () => {
    if (inviteInfo?.invite_link) {
      navigator.clipboard.writeText(inviteInfo.invite_link);
      message.success('邀请链接已复制');
    }
  };

  const handleClaimReward = async (taskId: number) => {
    try {
      await referralApi.claimTaskReward(taskId);
      message.success('奖励领取成功');
      loadData();
    } catch (error) {
      message.error('领取失败');
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <h2>拉新推广</h2>
      
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic 
              title="累计邀请" 
              value={stats.totalInvited} 
              prefix={<UserAddOutlined />} 
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic 
              title="有效邀请" 
              value={stats.validInvited} 
              prefix={<UserAddOutlined />} 
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic 
              title="累计奖励" 
              value={stats.totalRewards / 100} 
              precision={2}
              prefix="¥" 
            />
          </Card>
        </Col>
      </Row>

      {/* 邀请码区域 */}
      <Card style={{ marginBottom: 24 }}>
        {inviteInfo ? (
          <div style={{ textAlign: 'center' }}>
            <h3>我的邀请码</h3>
            <div style={{ fontSize: 48, fontWeight: 'bold', color: '#FFD700', margin: '16px 0' }}>
              {inviteInfo.invite_code}
            </div>
            <div style={{ marginBottom: 16 }}>
              <QRCode value={inviteInfo.invite_link} size={128} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'center', gap: 16 }}>
              <Button icon={<CopyOutlined />} onClick={handleCopyCode}>
                复制邀请码
              </Button>
              <Button type="primary" icon={<ShareAltOutlined />} onClick={handleCopyLink}>
                复制邀请链接
              </Button>
            </div>
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <p>您还没有邀请码，点击下方按钮生成</p>
            <Button type="primary" size="large" onClick={handleGenerateCode}>
              生成邀请码
            </Button>
          </div>
        )}
      </Card>

      <Tabs defaultActiveKey="tasks">
        <TabPane tab="任务奖励" key="tasks">
          <List
            loading={loading}
            dataSource={tasks}
            renderItem={(task) => (
              <List.Item
                actions={[
                  <Button 
                    type="primary" 
                    icon={<GiftOutlined />}
                    disabled={task.status !== 1}
                    onClick={() => handleClaimReward(task.id)}
                  >
                    {task.status === 1 ? '领取奖励' : '未完成'}
                  </Button>
                ]}
              >
                <List.Item.Meta
                  title={task.task_name}
                  description={
                    <div>
                      <p>{task.task_description}</p>
                      <Progress percent={Math.round((task.target_count / 10) * 100)} size="small" />
                      <Tag color="gold">奖励 ¥{(task.reward_amount / 100).toFixed(2)}</Tag>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        </TabPane>

        <TabPane tab="邀请记录" key="records">
          <List
            loading={loading}
            dataSource={inviteRecords}
            renderItem={(record) => (
              <List.Item>
                <List.Item.Meta
                  title={record.invited_user_name || record.invited_user_phone}
                  description={`邀请时间: ${record.created_at}`}
                />
                <div>
                  <Tag color={record.status === 1 ? 'success' : 'warning'}>
                    {record.status === 1 ? '有效' : '待激活'}
                  </Tag>
                  {record.reward_amount > 0 && (
                    <Tag color="gold">+¥{(record.reward_amount / 100).toFixed(2)}</Tag>
                  )}
                </div>
              </List.Item>
            )}
          />
        </TabPane>

        <TabPane tab="奖励记录" key="rewards">
          <List
            loading={loading}
            dataSource={rewards}
            renderItem={(record) => (
              <List.Item>
                <List.Item.Meta
                  title={record.reward_type === 'task' ? '任务奖励' : '邀请奖励'}
                  description={`发放时间: ${record.created_at}`}
                />
                <div>
                  <Tag color="success">+¥{(record.reward_amount / 100).toFixed(2)}</Tag>
                </div>
              </List.Item>
            )}
          />
        </TabPane>
      </Tabs>
    </div>
  );
};

export default ReferralPage;
