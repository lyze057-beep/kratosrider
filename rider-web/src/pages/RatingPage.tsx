import React, { useState, useEffect } from 'react';
import { Card, Table, Rate, Tag, Button, Modal, Form, Input, message, Statistic, Row, Col } from 'antd';
import { StarOutlined } from '@ant-design/icons';
import { ratingApi, RatingInfo, RatingSummary, RATING_SOURCE_TEXT } from '../api/rating';
import { useAuthStore } from '../store/authStore';

const RatingPage: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [ratings, setRatings] = useState<RatingInfo[]>([]);
  const [summary, setSummary] = useState<RatingSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [replyModalVisible, setReplyModalVisible] = useState(false);
  const [selectedRating, setSelectedRating] = useState<RatingInfo | null>(null);
  const [replyForm] = Form.useForm();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const riderId = Number(userInfo?.userId || 0);
      const [recordsResult, summaryResult] = await Promise.all([
        ratingApi.getRatingRecords({ rider_id: riderId, page: 1, page_size: 20 }),
        ratingApi.getRatingSummary(riderId)
      ]);
      setRatings(recordsResult.records || []);
      setSummary(summaryResult.summary);
    } catch (error) {
      console.error('Load rating data error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleReply = (rating: RatingInfo) => {
    setSelectedRating(rating);
    setReplyModalVisible(true);
  };

  const handleSubmitReply = async (values: any) => {
    if (!selectedRating) return;
    try {
      await ratingApi.replyToRating(selectedRating.id, values.content);
      message.success('回复成功');
      setReplyModalVisible(false);
      replyForm.resetFields();
      loadData();
    } catch (error) {
      message.error('回复失败');
    }
  };

  const columns = [
    {
      title: '评分',
      dataIndex: 'score',
      key: 'score',
      render: (score: number) => <Rate disabled defaultValue={score} />,
    },
    {
      title: '来源',
      dataIndex: 'source_type',
      key: 'source_type',
      render: (type: number) => <Tag>{RATING_SOURCE_TEXT[type] || '未知'}</Tag>,
    },
    {
      title: '内容',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: '回复',
      key: 'reply',
      render: (_: any, record: RatingInfo) => (
        record.reply_content ? (
          <Tag color="success">已回复</Tag>
        ) : (
          <Button type="link" onClick={() => handleReply(record)}>
            回复
          </Button>
        )
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card title="我的评价" style={{ marginBottom: 16 }}>
        {summary && (
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={6}>
              <Statistic
                title="平均评分"
                value={summary.average_score}
                precision={1}
                suffix="分"
                prefix={<StarOutlined />}
              />
            </Col>
            <Col span={6}>
              <Statistic title="总评价数" value={summary.total_count} />
            </Col>
            <Col span={6}>
              <Statistic title="五星好评" value={summary.five_star_count} />
            </Col>
            <Col span={6}>
              <Statistic title="好评率" value={summary.total_count > 0 ? ((summary.five_star_count + summary.four_star_count) / summary.total_count * 100).toFixed(1) : 0} suffix="%" />
            </Col>
          </Row>
        )}

        <Table
          columns={columns}
          dataSource={ratings}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="回复评价"
        open={replyModalVisible}
        onCancel={() => setReplyModalVisible(false)}
        footer={null}
      >
        <Form form={replyForm} onFinish={handleSubmitReply} layout="vertical">
          <Form.Item
            name="content"
            label="回复内容"
            rules={[{ required: true, message: '请输入回复内容' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入回复内容" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              提交回复
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default RatingPage;
