import { ComponentFixture, TestBed } from '@angular/core/testing';

import { KtbTriggerSequenceComponent } from './ktb-trigger-sequence.component';
import { AppModule } from '../../app.module';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { TRIGGER_EVALUATION_TIME, TRIGGER_SEQUENCE } from '../../_models/trigger-sequence';
import { Timeframe } from '../../_models/timeframe';
import moment from 'moment';
import { DataService } from '../../_services/data.service';

describe('KtbTriggerSequenceComponent', () => {
  let component: KtbTriggerSequenceComponent;
  let fixture: ComponentFixture<KtbTriggerSequenceComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppModule, HttpClientTestingModule],
    }).compileComponents();

    fixture = TestBed.createComponent(KtbTriggerSequenceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    component.projectName = 'podtato-head';
  });

  it('should create', () => {
    expect(component).toBeTruthy();
    expect(component.selectedStage).toEqual(undefined);
  });

  it('should set a given selected stage on init and not be changed afterwards', () => {
    // given, when
    fixture = TestBed.createComponent(KtbTriggerSequenceComponent);
    component = fixture.componentInstance;
    component.projectName = 'podtato-head';
    component.stage = 'hardening';
    fixture.detectChanges();

    // then
    expect(component.selectedStage).toEqual('hardening');

    // when
    component.stage = 'production';

    // then
    expect(component.selectedStage).toEqual('hardening');
  });

  it('should set the form state', () => {
    // given, when
    component.sequenceType = TRIGGER_SEQUENCE.EVALUATION;

    // when
    component.setFormState();

    // then
    expect(component.state).toEqual(TRIGGER_SEQUENCE.EVALUATION);
  });

  it('should test if a string is valid', () => {
    expect(component.isValidString(undefined)).toEqual(false);
    expect(component.isValidString('')).toEqual(false);
    expect(component.isValidString('    ')).toEqual(false);
    expect(component.isValidString('   test   ')).toEqual(true);
    expect(component.isValidString('test')).toEqual(true);
  });

  it('should test if start and end date are valid (start has to be before end)', () => {
    // given
    const start = moment().date(2).month(1).year(2021);
    const end = moment().date(1).month(1).year(2021);

    // when, then
    component.setStartDate(undefined);
    component.setEndDate(undefined);
    expect(component.isValidStartEndTime()).toEqual(false);
    component.setStartDate(undefined);
    component.setEndDate(end.toISOString());
    expect(component.isValidStartEndTime()).toEqual(false);
    component.setStartDate(start.toISOString());
    component.setEndDate(undefined);
    expect(component.isValidStartEndTime()).toEqual(false);

    component.setStartDate(start.toISOString());
    component.setEndDate(end.toISOString());
    expect(component.isValidStartBeforeEnd).toEqual(false);
    expect(component.isValidStartEndTime()).toEqual(false);

    start.hours(12).minutes(0).seconds(0);
    end.hours(12).minutes(0).seconds(5);
    component.setStartDate(start.toISOString());
    component.setEndDate(end.toISOString());
    expect(component.isValidStartEndDuration).toEqual(false);
    expect(component.isValidStartEndTime()).toEqual(false);

    end.date(3).month(1).year(2021);
    component.setStartDate(start.toISOString());
    component.setEndDate(end.toISOString());
    expect(component.isValidStartEndDuration).toEqual(true);
    expect(component.isValidStartBeforeEnd).toEqual(true);
    expect(component.isValidStartEndTime()).toEqual(true);
  });

  it('should return an combined and cleaned string for image and tag', () => {
    // given
    const image = '  docker  .io/keptn  ';
    const tag = '  v0.1 . 2    ';

    // when, then
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    expect(component.getImageString(image, tag)).toEqual('docker.io/keptn:v0.1.2');
  });

  it('should parse a timeframe to a string in format 1h1m1s1ms1us', () => {
    // given
    const timeframe: Timeframe = {
      hours: 1,
      minutes: undefined,
      seconds: undefined,
      millis: undefined,
      micros: undefined,
    };

    // when, then
    /* eslint-disable @typescript-eslint/ban-ts-comment */
    // @ts-ignore
    expect(component.parseTimeframe(timeframe)).toEqual('1h');
    timeframe.hours = undefined;
    timeframe.minutes = 1;
    // @ts-ignore
    expect(component.parseTimeframe(timeframe)).toEqual('1m');
    timeframe.minutes = undefined;
    timeframe.seconds = 1;
    // @ts-ignore
    expect(component.parseTimeframe(timeframe)).toEqual('1s');
    timeframe.seconds = undefined;
    timeframe.millis = 1;
    // @ts-ignore
    expect(component.parseTimeframe(timeframe)).toEqual('1ms');
    timeframe.millis = undefined;
    timeframe.micros = 1;
    // @ts-ignore
    expect(component.parseTimeframe(timeframe)).toEqual('1us');
    timeframe.hours = 1;
    timeframe.minutes = 1;
    timeframe.seconds = 1;
    timeframe.millis = 1;
    timeframe.micros = 1;
    // @ts-ignore
    expect(component.parseTimeframe(timeframe)).toEqual('1h1m1s1ms1us');
    /* eslint-enable */
  });

  it('should be a valid timeframe', () => {
    // given
    const timeframe: Timeframe = {
      hours: undefined,
      minutes: undefined,
      seconds: undefined,
      millis: undefined,
      micros: undefined,
    };

    assertTimeframeValid(timeframe, true);

    timeframe.hours = 1;
    assertTimeframeValid(timeframe, true);

    timeframe.hours = undefined;
    timeframe.minutes = 1;
    assertTimeframeValid(timeframe, true);

    timeframe.minutes = undefined;
    timeframe.seconds = 60;
    assertTimeframeValid(timeframe, true);

    timeframe.seconds = undefined;
    timeframe.millis = 60_000;
    assertTimeframeValid(timeframe, true);

    timeframe.millis = undefined;
    timeframe.micros = 60_000_000;
    assertTimeframeValid(timeframe, true);
  });

  it('should be an invalid timeframe', () => {
    // given
    const timeframe: Timeframe = {
      hours: 0,
      minutes: undefined,
      seconds: undefined,
      millis: undefined,
      micros: undefined,
    };

    assertTimeframeValid(timeframe, false);

    timeframe.hours = undefined;
    timeframe.minutes = 0;
    assertTimeframeValid(timeframe, false);

    timeframe.minutes = undefined;
    timeframe.seconds = 59;
    assertTimeframeValid(timeframe, false);

    timeframe.seconds = undefined;
    timeframe.millis = 59_999;
    assertTimeframeValid(timeframe, false);

    timeframe.millis = undefined;
    timeframe.micros = 59_999_999;
    assertTimeframeValid(timeframe, false);
  });

  it('should parse labels to an object', () => {
    // given
    let labels = '   key1 = val1, value2   ,key2   = val3           ';

    // when, then
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    expect(component.parseLabels(labels)).toEqual({ key1: 'val1', key2: 'val3' });

    // given
    labels = '';

    // when, then
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    expect(component.parseLabels(labels)).toEqual({});
  });

  it('should trigger the delivery sequence with set data', () => {
    // given
    const dataService = TestBed.inject(DataService);
    const spy = jest.spyOn(dataService, 'triggerDelivery');
    component.sequenceType = TRIGGER_SEQUENCE.DELIVERY;
    component.selectedStage = 'hardening';
    component.selectedService = 'helloservice';
    component.deliveryFormData = {
      image: 'docker.io/keptn',
      tag: 'v0.1.2',
      labels: 'key1=val1',
      values: '{"key2": "val2", "key3": {"key4": "val3"}}',
    };

    // when
    component.triggerSequence();

    // then
    expect(spy).toHaveBeenCalledWith({
      project: 'podtato-head',
      stage: 'hardening',
      service: 'helloservice',
      labels: {
        key1: 'val1',
      },
      configurationChange: {
        values: {
          key2: 'val2',
          key3: {
            key4: 'val3',
          },
          image: 'docker.io/keptn:v0.1.2',
        },
      },
    });
  });

  it('should trigger the evaluation sequence with a timeframe set', () => {
    // given
    const dataService = TestBed.inject(DataService);
    const spy = jest.spyOn(dataService, 'triggerEvaluation');
    const date = moment();
    component.sequenceType = TRIGGER_SEQUENCE.EVALUATION;
    component.selectedStage = 'hardening';
    component.selectedService = 'helloservice';
    component.evaluationFormData = {
      evaluationType: TRIGGER_EVALUATION_TIME.TIMEFRAME,
      labels: 'key1=val1',
      timeframe: {
        hours: 1,
        minutes: 15,
        seconds: undefined,
        millis: undefined,
        micros: undefined,
      },
      timeframeStart: date.toISOString(),
      startDatetime: undefined,
      endDatetime: undefined,
    };

    // when
    component.triggerSequence();

    // then
    expect(spy).toHaveBeenCalledWith({
      project: 'podtato-head',
      stage: 'hardening',
      service: 'helloservice',
      evaluation: {
        labels: {
          key1: 'val1',
        },
        timeframe: '1h15m',
        start: date.toISOString(),
      },
    });
  });

  it('should trigger the evaluation sequence with a start / end date set', () => {
    // given
    const dataService = TestBed.inject(DataService);
    const spy = jest.spyOn(dataService, 'triggerEvaluation');
    const start = moment().date(1).month(1).year(2021);
    const end = moment().date(2).month(1).year(2021);
    component.sequenceType = TRIGGER_SEQUENCE.EVALUATION;
    component.selectedStage = 'hardening';
    component.selectedService = 'helloservice';
    component.evaluationFormData = {
      evaluationType: TRIGGER_EVALUATION_TIME.START_END,
      labels: 'key1=val1',
      timeframe: undefined,
      timeframeStart: undefined,
      startDatetime: start.toISOString(),
      endDatetime: end.toISOString(),
    };

    // when
    component.triggerSequence();

    // then
    expect(spy).toHaveBeenCalledWith({
      project: 'podtato-head',
      stage: 'hardening',
      service: 'helloservice',
      evaluation: {
        labels: {
          key1: 'val1',
        },
        start: start.toISOString(),
        end: end.toISOString(),
      },
    });
  });

  it('should trigger the evaluation sequence with no timeframe set', () => {
    // given
    const dataService = TestBed.inject(DataService);
    const spy = jest.spyOn(dataService, 'triggerEvaluation');
    const date = moment().milliseconds(0);
    component.sequenceType = TRIGGER_SEQUENCE.EVALUATION;
    component.selectedStage = 'hardening';
    component.selectedService = 'helloservice';
    component.evaluationFormData = {
      evaluationType: TRIGGER_EVALUATION_TIME.TIMEFRAME,
      labels: 'key1=val1',
      timeframe: {
        hours: undefined,
        minutes: undefined,
        seconds: undefined,
        millis: undefined,
        micros: undefined,
      },
      timeframeStart: date.toISOString(),
      startDatetime: undefined,
      endDatetime: undefined,
    };

    // when
    component.triggerSequence();

    // then
    expect(spy).toHaveBeenCalledWith({
      project: 'podtato-head',
      stage: 'hardening',
      service: 'helloservice',
      evaluation: {
        labels: {
          key1: 'val1',
        },
        timeframe: '5m',
        start: date.toISOString(),
      },
    });
  });

  it('should trigger the custom sequence with set data', () => {
    // given
    const dataService = TestBed.inject(DataService);
    const spy = jest.spyOn(dataService, 'triggerCustomSequence');
    component.sequenceType = TRIGGER_SEQUENCE.CUSTOM;
    component.selectedStage = 'hardening';
    component.selectedService = 'helloservice';
    component.customFormData = {
      project: 'podtato-head',
      stage: 'hardening',
      service: 'helloservice',
      sequence: 'testsequence',
      labels: 'key1=val1',
    };

    // when
    component.triggerSequence();

    // then
    expect(spy).toHaveBeenCalledWith(
      {
        project: 'podtato-head',
        stage: 'hardening',
        service: 'helloservice',
        labels: {
          key1: 'val1',
        },
      },
      'testsequence'
    );
  });

  function assertTimeframeValid(timeframe: Timeframe, isValid: boolean): void {
    component.setTimeframe(timeframe);
    expect(component.isValidTimeframe).toEqual(isValid);
  }
});
