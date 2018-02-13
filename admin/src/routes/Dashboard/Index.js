import React, { Component } from 'react';
import { Row, Col, Card } from 'antd';

export default class Dashboard extends Component {
  render() {
    const topColResponsiveProps = {
      xs: 24,
      sm: 12,
      md: 12,
      lg: 12,
      xl: 6,
      style: { marginBottom: 24 },
    };

    return (
      <Row gutter={24}>
        <Col {...topColResponsiveProps}>
          <Card
            title="今日新增用户"
          >
            <h1>100</h1>
          </Card>
        </Col>
        <Col {...topColResponsiveProps}>
          <Card
            title="今日新增主题"
          >
            <h1>aa</h1>
          </Card>
        </Col>
        <Col {...topColResponsiveProps}>
          <Card
            title="今日新增评论"
          >
            <h1>bb</h1>
          </Card>
        </Col>
        <Col {...topColResponsiveProps}>
          <Card
            title="用户总数"
          >
            <h1>199</h1>
          </Card>
        </Col>
      </Row>
    );
  }
}
